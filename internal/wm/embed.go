package wm

import (
	"fmt"

	spectralimage "spectralmark/internal/image"
	spectralmath "spectralmark/internal/math"
)

type coeffPos struct {
	u int
	v int
}

var midFreqPositions = [...]coeffPos{
	{u: 1, v: 2},
	{u: 2, v: 1},
	{u: 2, v: 2},
	{u: 3, v: 1},
	{u: 1, v: 3},
	{u: 3, v: 2},
}

func EmbedPPM(inPath, outPath, key, msg string, alpha float32) error {
	if inPath == "" {
		return fmt.Errorf("input path is required")
	}
	if outPath == "" {
		return fmt.Errorf("output path is required")
	}
	if key == "" {
		return fmt.Errorf("key is required")
	}
	if alpha <= 0 {
		return fmt.Errorf("alpha must be > 0")
	}

	img, err := spectralimage.ReadPPM(inPath)
	if err != nil {
		return err
	}

	y, cb, cr := spectralimage.RGBToYCbCr(img)
	yPad, w2, h2 := spectralmath.PadTo8(y, img.W, img.H)

	bits := EncodePayload(msg)
	blockCols := w2 / 8
	blockRows := h2 / 8
	capacity := blockCols * blockRows * len(midFreqPositions)
	if len(bits) > capacity {
		return fmt.Errorf("payload too large for image: payload symbols=%d capacity=%d", len(bits), capacity)
	}

	rng := NewPRNG(SeedFromKey(key))
	bitIdx := 0

embedLoop:
	for by := 0; by < blockRows; by++ {
		for bx := 0; bx < blockCols; bx++ {
			if bitIdx >= len(bits) {
				break embedLoop
			}

			block := spectralmath.GetBlock8(yPad, w2, bx, by)
			coeff := spectralmath.DCT8(block)

			for _, pos := range midFreqPositions {
				if bitIdx >= len(bits) {
					break
				}

				pn := float32(1)
				if rng.NextPM1() < 0 {
					pn = -1
				}

				direction := float32(bits[bitIdx]) * pn
				projected := coeff[pos.v][pos.u] * direction
				target := alpha * 0.7
				if projected < target {
					coeff[pos.v][pos.u] += (target - projected) * direction
				}
				bitIdx++
			}

			recon := spectralmath.IDCT8(coeff)
			clampBlockToByteRange(&recon)
			spectralmath.SetBlock8(yPad, w2, bx, by, recon)
		}
	}

	yOut := spectralmath.Unpad(yPad, w2, h2, img.W, img.H)
	outImg := spectralimage.YCbCrToRGB(img.W, img.H, yOut, cb, cr)

	return spectralimage.WritePPM(outPath, outImg)
}

func clampBlockToByteRange(b *[8][8]float32) {
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if b[y][x] < 0 {
				b[y][x] = 0
				continue
			}
			if b[y][x] > 255 {
				b[y][x] = 255
			}
		}
	}
}
