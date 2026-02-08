package wm

const (
	payloadSyncWord  = 0xA55A
	maxPayloadBytes  = 0xFFFF
	repetitionFactor = 3
)

func CRC16(data []byte) uint16 {
	// CRC-16/CCITT-FALSE: poly 0x1021, init 0xFFFF.
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

func EncodePayload(msg string) []int8 {
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

	out := make([]int8, 0, len(rawBits)*repetitionFactor)
	for _, bit := range rawBits {
		symbol := int8(-1)
		if bit == 1 {
			symbol = 1
		}
		for i := 0; i < repetitionFactor; i++ {
			out = append(out, symbol)
		}
	}

	return out
}

func DecodePayload(bits []int8) (msg string, ok bool) {
	if len(bits) < (16+16+16)*repetitionFactor {
		return "", false
	}

	rawBits := majorityDecodeBits(bits)
	if len(rawBits) < 48 {
		return "", false
	}

	for start := 0; start+32 <= len(rawBits); start++ {
		sync := readWordAtBit(rawBits, start)
		if sync != payloadSyncWord {
			continue
		}

		msgLen := int(readWordAtBit(rawBits, start+16))
		totalNeeded := start + 16 + 16 + msgLen*8 + 16
		if totalNeeded > len(rawBits) {
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

		return string(data), true
	}

	return "", false
}

func majorityDecodeBits(bits []int8) []uint8 {
	n := len(bits) / repetitionFactor
	out := make([]uint8, n)
	for i := 0; i < n; i++ {
		votes := 0
		base := i * repetitionFactor
		for j := 0; j < repetitionFactor; j++ {
			if bits[base+j] > 0 {
				votes++
			} else {
				votes--
			}
		}
		if votes > 0 {
			out[i] = 1
		}
	}
	return out
}

func appendWordBits(dst []uint8, v uint16) []uint8 {
	dst = appendByteBits(dst, byte(v>>8))
	dst = appendByteBits(dst, byte(v))
	return dst
}

func appendByteBits(dst []uint8, b byte) []uint8 {
	for i := 7; i >= 0; i-- {
		dst = append(dst, uint8((b>>i)&1))
	}
	return dst
}

func readWordAtBit(bits []uint8, start int) uint16 {
	hi := readByteAtBit(bits, start)
	lo := readByteAtBit(bits, start+8)
	return uint16(hi)<<8 | uint16(lo)
}

func readByteAtBit(bits []uint8, start int) byte {
	var out byte
	for i := 0; i < 8; i++ {
		out <<= 1
		idx := start + i
		if idx >= 0 && idx < len(bits) && bits[idx] != 0 {
			out |= 1
		}
	}
	return out
}
