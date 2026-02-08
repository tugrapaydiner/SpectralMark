package math

import stdmath "math"

const dctSize = 8

var (
	dctAlpha = [dctSize]float32{
		float32(1.0 / stdmath.Sqrt2),
		1, 1, 1, 1, 1, 1, 1,
	}
	dctCos = buildDCTCosTable()
)

func DCT8(block [8][8]float32) [8][8]float32 {
	var coeff [8][8]float32

	for v := 0; v < dctSize; v++ {
		for u := 0; u < dctSize; u++ {
			sum := float32(0)
			for y := 0; y < dctSize; y++ {
				for x := 0; x < dctSize; x++ {
					sum += block[y][x] * dctCos[u][x] * dctCos[v][y]
				}
			}
			coeff[v][u] = 0.25 * dctAlpha[u] * dctAlpha[v] * sum
		}
	}

	return coeff
}

func IDCT8(coeff [8][8]float32) [8][8]float32 {
	var block [8][8]float32

	for y := 0; y < dctSize; y++ {
		for x := 0; x < dctSize; x++ {
			sum := float32(0)
			for v := 0; v < dctSize; v++ {
				for u := 0; u < dctSize; u++ {
					sum += dctAlpha[u] * dctAlpha[v] * coeff[v][u] * dctCos[u][x] * dctCos[v][y]
				}
			}
			block[y][x] = 0.25 * sum
		}
	}

	return block
}

func buildDCTCosTable() [dctSize][dctSize]float32 {
	var table [dctSize][dctSize]float32
	for u := 0; u < dctSize; u++ {
		for x := 0; x < dctSize; x++ {
			angle := ((2*float64(x) + 1) * float64(u) * stdmath.Pi) / 16.0
			table[u][x] = float32(stdmath.Cos(angle))
		}
	}
	return table
}
