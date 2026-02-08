package wm

import (
	"fmt"

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

	rng := NewPRNG(SeedFromKey(key))
	softSymbols := make([]float32, 0, blockCols*blockRows*len(midFreqPositions))

	for by := 0; by < blockRows; by++ {
		for bx := 0; bx < blockCols; bx++ {
			block := spectralmath.GetBlock8(yPad, w2, bx, by)
			coeff := spectralmath.DCT8(block)

			for _, pos := range midFreqPositions {
				pn := float32(1)
				if rng.NextPM1() < 0 {
					pn = -1
				}

				softSymbols = append(softSymbols, coeff[pos.v][pos.u]*pn)
			}
		}
	}

	if len(softSymbols) == 0 {
		return 0, false, "", false, nil
	}

	rawBits := decodeRepetitionSoft(softSymbols)
	msg, ok, startBit, payloadBits := decodePayloadFromRawBits(rawBits)
	score = estimateDetectScore(rawBits, msg, ok, startBit, payloadBits)
	present = ok

	return
}

func decodeRepetitionSoft(softSymbols []float32) []uint8 {
	n := len(softSymbols) / repetitionFactor
	out := make([]uint8, n)
	for i := 0; i < n; i++ {
		base := i * repetitionFactor
		sum := softSymbols[base] + softSymbols[base+1] + softSymbols[base+2]
		if sum >= 0 {
			out[i] = 1
		}
	}
	return out
}

func decodePayloadFromRawBits(rawBits []uint8) (msg string, ok bool, startBit int, payloadBits int) {
	if len(rawBits) < 48 {
		return "", false, -1, 0
	}

	for start := 0; start+32 <= len(rawBits); start++ {
		sync := readWordAtBit(rawBits, start)
		if sync != payloadSyncWord {
			continue
		}

		msgLen := int(readWordAtBit(rawBits, start+16))
		totalNeeded := 16 + 16 + msgLen*8 + 16
		if start+totalNeeded > len(rawBits) {
			continue
		}

		data := make([]byte, msgLen)
		dataStart := start + 32
		for i := 0; i < msgLen; i++ {
			data[i] = readByteAtBit(rawBits, dataStart+i*8)
		}

		gotCRC := readWordAtBit(rawBits, dataStart+msgLen*8)
		if gotCRC != CRC16(data) {
			continue
		}

		return string(data), true, start, totalNeeded
	}

	return "", false, -1, 0
}

func estimateDetectScore(rawBits []uint8, msg string, ok bool, startBit int, payloadBits int) float32 {
	if len(rawBits) == 0 {
		return 0
	}

	if ok {
		expected := encodePayloadRawBits(msg)
		n := len(expected)
		if payloadBits > 0 && payloadBits < n {
			n = payloadBits
		}
		if startBit < 0 || startBit >= len(rawBits) {
			return 0
		}
		if startBit+n > len(rawBits) {
			n = len(rawBits) - startBit
		}
		if n == 0 {
			return 0
		}

		matches := 0
		for i := 0; i < n; i++ {
			if rawBits[startBit+i] == expected[i] {
				matches++
			}
		}
		return float32(matches) / float32(n)
	}

	if len(rawBits) < 16 {
		return 0
	}

	syncBits := appendWordBits(nil, payloadSyncWord)
	matches := 0
	for i := 0; i < 16; i++ {
		if rawBits[i] == syncBits[i] {
			matches++
		}
	}

	return float32(matches) / 16.0
}

func encodePayloadRawBits(msg string) []uint8 {
	data := []byte(msg)
	if len(data) > maxPayloadBytes {
		data = data[:maxPayloadBytes]
	}

	rawBits := make([]uint8, 0, (2+2+len(data)+2)*8)
	rawBits = appendWordBits(rawBits, uint16(payloadSyncWord))
	rawBits = appendWordBits(rawBits, uint16(len(data)))
	for _, b := range data {
		rawBits = appendByteBits(rawBits, b)
	}
	rawBits = appendWordBits(rawBits, CRC16(data))
	return rawBits
}
