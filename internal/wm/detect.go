package wm

import (
	"fmt"
	"sort"

	spectralimage "spectralmark/internal/image"
	spectralmath "spectralmark/internal/math"
)

func DetectPPM(path, key string) (score float32, present bool, msg string, ok bool, err error) {
	if path == "" {
		err = fmt.Errorf("input path is required")
		return
	}
	if key == "" {
		err = fmt.Errorf("key is required")
		return
	}

	img, readErr := spectralimage.ReadPPM(path)
	if readErr != nil {
		err = readErr
		return
	}

	y, _, _ := spectralimage.RGBToYCbCr(img)
	yPad, w2, h2 := spectralmath.PadTo8(y, img.W, img.H)
	if w2 <= 0 || h2 <= 0 {
		return 0, false, "", false, nil
	}

	blockCols := w2 / 8
	blockRows := h2 / 8
	blockCount := blockCols * blockRows
	if blockCount <= 0 {
		return 0, false, "", false, nil
	}
	totalSlots := blockCount * len(midFreqPositions)
	if totalSlots < spreadChipsPerSymbol {
		return 0, false, "", false, nil
	}

	coeffVals := make([][]float32, blockCount)
	for blockIdx := 0; blockIdx < blockCount; blockIdx++ {
		bx := blockIdx % blockCols
		by := blockIdx / blockCols

		block := spectralmath.GetBlock8(yPad, w2, bx, by)
		coeff := spectralmath.DCT8(block)

		row := make([]float32, len(midFreqPositions))
		for i, pos := range midFreqPositions {
			row[i] = coeff[pos.v][pos.u]
		}
		coeffVals[blockIdx] = row
	}

	symbolCount := totalSlots / spreadChipsPerSymbol
	neededSlots := symbolCount * spreadChipsPerSymbol
	slots, chips := shuffledSlotsAndChips(key, totalSlots, neededSlots)
	if len(slots) != neededSlots || len(chips) != neededSlots {
		return 0, false, "", false, nil
	}

	symbolSoft := make([]float32, symbolCount)
	symbols := make([]int8, symbolCount)
	for symIdx := 0; symIdx < symbolCount; symIdx++ {
		soft := float32(0)
		base := symIdx * spreadChipsPerSymbol
		for j := 0; j < spreadChipsPerSymbol; j++ {
			slotIdx := base + j
			slot := slots[slotIdx]
			blockIdx := slot / len(midFreqPositions)
			coeffIdx := slot % len(midFreqPositions)

			soft += coeffVals[blockIdx][coeffIdx] * float32(chips[slotIdx])
		}
		symbolSoft[symIdx] = soft
		if soft >= 0 {
			symbols[symIdx] = 1
		} else {
			symbols[symIdx] = -1
		}
	}

	if len(symbols) == 0 {
		return 0, false, "", false, nil
	}

	msg, ok = decodePayloadFromSymbolSoft(symbolSoft, 2, 8)
	score = estimateDetectScoreSymbols(symbols, msg, ok)
	present = ok

	return
}

func decodePayloadFromSymbolSoft(symbolSoft []float32, maxSyncErrors int, maxDataCandidates int) (msg string, ok bool) {
	rawBits, rawConf := repetitionSoftToRaw(symbolSoft)
	if len(rawBits) < 48 {
		return "", false
	}

	// First pass: direct CRC check with tolerant sync matching.
	if msg, ok = decodeRawBitsDirect(rawBits, maxSyncErrors); ok {
		return msg, true
	}

	// Second pass: flip low-confidence data bits (not sync/len/crc), then re-check CRC.
	return decodeRawBitsWithBitFixes(rawBits, rawConf, maxSyncErrors, maxDataCandidates)
}

func repetitionSoftToRaw(symbolSoft []float32) (rawBits []uint8, rawConf []float32) {
	rawCount := len(symbolSoft) / repetitionFactor
	rawBits = make([]uint8, rawCount)
	rawConf = make([]float32, rawCount)

	for i := 0; i < rawCount; i++ {
		base := i * repetitionFactor
		sum := symbolSoft[base] + symbolSoft[base+1] + symbolSoft[base+2]
		if sum >= 0 {
			rawBits[i] = 1
			rawConf[i] = sum
		} else {
			rawConf[i] = -sum
		}
	}

	return rawBits, rawConf
}

func decodeRawBitsDirect(rawBits []uint8, maxSyncErrors int) (msg string, ok bool) {
	if len(rawBits) < 48 {
		return "", false
	}

	syncBits := appendWordBits(nil, payloadSyncWord)
	for start := 0; start+32 <= len(rawBits); start++ {
		errCount := 0
		for i := 0; i < 16; i++ {
			if rawBits[start+i] != syncBits[i] {
				errCount++
				if errCount > maxSyncErrors {
					break
				}
			}
		}
		if errCount > maxSyncErrors {
			continue
		}

		fieldLen := readWordAtBit(rawBits, start+16)
		maxLenFit := (len(rawBits) - start - 32 - 16) / 8
		if maxLenFit < 0 {
			continue
		}

		for _, msgLen := range candidateLengths(fieldLen, maxLenFit) {
			totalNeeded := start + 16 + 16 + msgLen*8 + 16
			if msgLen < 0 || totalNeeded > len(rawBits) {
				continue
			}

			dataStart := start + 32
			data := make([]byte, msgLen)
			for i := 0; i < msgLen; i++ {
				data[i] = readByteAtBit(rawBits, dataStart+i*8)
			}

			gotCRC := readWordAtBit(rawBits, dataStart+msgLen*8)
			if gotCRC != CRC16(data) {
				continue
			}

			return string(data), true
		}
	}

	return "", false
}

func decodeRawBitsWithBitFixes(rawBits []uint8, rawConf []float32, maxSyncErrors int, maxDataCandidates int) (msg string, ok bool) {
	if len(rawBits) < 48 {
		return "", false
	}
	if maxDataCandidates <= 0 {
		return "", false
	}

	syncBits := appendWordBits(nil, payloadSyncWord)
	for start := 0; start+32 <= len(rawBits); start++ {
		errCount := 0
		for i := 0; i < 16; i++ {
			if rawBits[start+i] != syncBits[i] {
				errCount++
				if errCount > maxSyncErrors {
					break
				}
			}
		}
		if errCount > maxSyncErrors {
			continue
		}

		fieldLen := readWordAtBit(rawBits, start+16)
		maxLenFit := (len(rawBits) - start - 32 - 16) / 8
		if maxLenFit < 0 {
			continue
		}

		for _, msgLen := range candidateLengths(fieldLen, maxLenFit) {
			totalNeeded := start + 16 + 16 + msgLen*8 + 16
			if msgLen < 0 || totalNeeded > len(rawBits) {
				continue
			}

			dataBitStart := start + 32
			dataBitCount := msgLen * 8
			if dataBitCount <= 0 {
				continue
			}
			crcStart := dataBitStart + dataBitCount
			gotCRC := readWordAtBit(rawBits, crcStart)

			candidateIdx := make([]int, 0, dataBitCount)
			for i := 0; i < dataBitCount; i++ {
				candidateIdx = append(candidateIdx, dataBitStart+i)
			}
			sort.Slice(candidateIdx, func(i, j int) bool {
				return rawConf[candidateIdx[i]] < rawConf[candidateIdx[j]]
			})
			if len(candidateIdx) > maxDataCandidates {
				candidateIdx = candidateIdx[:maxDataCandidates]
			}

			// Try single-bit corrections.
			for _, idx := range candidateIdx {
				testBits := make([]uint8, len(rawBits))
				copy(testBits, rawBits)
				testBits[idx] ^= 1

				data := make([]byte, msgLen)
				for i := 0; i < msgLen; i++ {
					data[i] = readByteAtBit(testBits, dataBitStart+i*8)
				}
				if CRC16(data) == gotCRC {
					return string(data), true
				}
			}

			// Try two-bit corrections.
			for i := 0; i < len(candidateIdx); i++ {
				for j := i + 1; j < len(candidateIdx); j++ {
					testBits := make([]uint8, len(rawBits))
					copy(testBits, rawBits)
					testBits[candidateIdx[i]] ^= 1
					testBits[candidateIdx[j]] ^= 1

					data := make([]byte, msgLen)
					for k := 0; k < msgLen; k++ {
						data[k] = readByteAtBit(testBits, dataBitStart+k*8)
					}
					if CRC16(data) == gotCRC {
						return string(data), true
					}
				}
			}

			// Try three-bit corrections.
			for i := 0; i < len(candidateIdx); i++ {
				for j := i + 1; j < len(candidateIdx); j++ {
					for m := j + 1; m < len(candidateIdx); m++ {
						testBits := make([]uint8, len(rawBits))
						copy(testBits, rawBits)
						testBits[candidateIdx[i]] ^= 1
						testBits[candidateIdx[j]] ^= 1
						testBits[candidateIdx[m]] ^= 1

						data := make([]byte, msgLen)
						for k := 0; k < msgLen; k++ {
							data[k] = readByteAtBit(testBits, dataBitStart+k*8)
						}
						if CRC16(data) == gotCRC {
							return string(data), true
						}
					}
				}
			}

		}
	}

	return "", false
}

func candidateLengths(fieldLen uint16, maxLenFit int) []int {
	if maxLenFit < 0 {
		return nil
	}

	out := make([]int, 0, maxLenFit+1)
	added := make([]bool, maxLenFit+1)

	if int(fieldLen) >= 0 && int(fieldLen) <= maxLenFit {
		out = append(out, int(fieldLen))
		added[int(fieldLen)] = true
	}

	for l := 0; l <= maxLenFit; l++ {
		if added[l] {
			continue
		}
		if hamming16(fieldLen, uint16(l)) <= 2 {
			out = append(out, l)
			added[l] = true
		}
	}

	return out
}

func hamming16(a, b uint16) int {
	x := a ^ b
	c := 0
	for x != 0 {
		x &= x - 1
		c++
	}
	return c
}

func estimateDetectScoreSymbols(symbols []int8, msg string, ok bool) float32 {
	if len(symbols) == 0 {
		return 0
	}

	if ok {
		expected := EncodePayload(msg)
		n := len(expected)
		if n > len(symbols) {
			n = len(symbols)
		}
		if n == 0 {
			return 0
		}

		matches := 0
		for i := 0; i < n; i++ {
			if symbols[i] == expected[i] {
				matches++
			}
		}
		return float32(matches) / float32(n)
	}

	logical := majorityDecodeBits(symbols)
	if len(logical) < 16 {
		return 0
	}

	syncBits := appendWordBits(nil, payloadSyncWord)
	matches := 0
	for i := 0; i < 16; i++ {
		if logical[i] == syncBits[i] {
			matches++
		}
	}

	return float32(matches) / 16.0
}
