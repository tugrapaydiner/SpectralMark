package util

import stdmath "math"

func PSNR(yA, yB []float32) float32 {
	if len(yA) == 0 || len(yA) != len(yB) {
		return 0
	}

	mse := float64(0)
	for i := range yA {
		d := float64(yA[i] - yB[i])
		mse += d * d
	}
	mse /= float64(len(yA))

	if mse == 0 {
		return float32(stdmath.Inf(1))
	}

	maxI := 255.0
	return float32(10.0 * stdmath.Log10((maxI*maxI)/mse))
}
