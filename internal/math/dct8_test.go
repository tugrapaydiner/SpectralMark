package math

import stdmath "math"
import "testing"

func TestDCT8IDCT8RoundTripErrorIsSmall(t *testing.T) {
	var block [8][8]float32
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			block[y][x] = float32((y*8+x)%17 - 8)
		}
	}

	coeff := DCT8(block)
	recon := IDCT8(coeff)

	maxErr := float32(0)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			err := float32(stdmath.Abs(float64(recon[y][x] - block[y][x])))
			if err > maxErr {
				maxErr = err
			}
		}
	}

	if maxErr > 1e-3 {
		t.Fatalf("max reconstruction error too high: got=%f want<=%f", maxErr, float32(1e-3))
	}
}

func TestDCT8ConstantBlockHasOnlyDC(t *testing.T) {
	var block [8][8]float32
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			block[y][x] = 10
		}
	}

	coeff := DCT8(block)
	if coeff[0][0] == 0 {
		t.Fatalf("expected non-zero DC term")
	}

	for v := 0; v < 8; v++ {
		for u := 0; u < 8; u++ {
			if u == 0 && v == 0 {
				continue
			}
			if abs32(coeff[v][u]) > 1e-3 {
				t.Fatalf("expected near-zero AC term at (%d,%d), got=%f", v, u, coeff[v][u])
			}
		}
	}
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
