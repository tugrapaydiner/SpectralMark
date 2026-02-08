package wm

import "testing"

func TestCRC16KnownVector(t *testing.T) {
	got := CRC16([]byte("123456789"))
	const want = uint16(0x29B1)
	if got != want {
		t.Fatalf("CRC16 mismatch: got=0x%04X want=0x%04X", got, want)
	}
}

func TestEncodeDecodePayloadRoundTrip(t *testing.T) {
	msg := "HELLO"
	bits := EncodePayload(msg)

	decoded, ok := DecodePayload(bits)
	if !ok {
		t.Fatalf("DecodePayload returned ok=false")
	}
	if decoded != msg {
		t.Fatalf("decoded mismatch: got=%q want=%q", decoded, msg)
	}
}

func TestDecodePayloadCorrectsSingleErrorPerTriplet(t *testing.T) {
	msg := "HELLO"
	bits := EncodePayload(msg)

	triplets := len(bits) / repetitionFactor
	for i := 0; i < triplets; i++ {
		// Flip one symbol in each triplet: majority vote should recover.
		bits[i*repetitionFactor] *= -1
	}

	decoded, ok := DecodePayload(bits)
	if !ok {
		t.Fatalf("DecodePayload returned ok=false")
	}
	if decoded != msg {
		t.Fatalf("decoded mismatch: got=%q want=%q", decoded, msg)
	}
}

func TestDecodePayloadRejectsCrcMismatch(t *testing.T) {
	msg := "HELLO"
	bits := EncodePayload(msg)

	// Corrupt one logical bit by flipping two symbols in the same triplet.
	if len(bits) < 6 {
		t.Fatalf("unexpected short encoded payload")
	}
	bits[0] *= -1
	bits[1] *= -1

	_, ok := DecodePayload(bits)
	if ok {
		t.Fatalf("expected decode failure from CRC mismatch")
	}
}
