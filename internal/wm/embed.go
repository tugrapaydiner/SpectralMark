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
	blockCount := blockCols * blockRows
	totalSlots := blockCount * len(midFreqPositions)
	neededSlots := len(bits) * spreadChipsPerSymbol
	if neededSlots > totalSlots {
		maxSymbols := totalSlots / spreadChipsPerSymbol
		return fmt.Errorf(
			"payload too large for image: payload symbols=%d capacity=%d (spread=%d)",
			len(bits),
			maxSymbols,
			spreadChipsPerSymbol,
		)
	}

	slots, chips := shuffledSlotsAndChips(key, totalSlots, neededSlots)
	if len(slots) != neededSlots || len(chips) != neededSlots {
		return fmt.Errorf("failed to allocate spread mapping")
	}

	type embedOp struct {
		coeffIdx  int
		direction float32
	}
	blockOps := make([][]embedOp, blockCount)

	for i := 0; i < neededSlots; i++ {
		symbolIdx := i / spreadChipsPerSymbol
		slot := slots[i]
		blockIdx := slot / len(midFreqPositions)
		coeffIdx := slot % len(midFreqPositions)

		direction := float32(bits[symbolIdx] * chips[i])
		blockOps[blockIdx] = append(blockOps[blockIdx], embedOp{
			coeffIdx:  coeffIdx,
			direction: direction,
		})
	}

	for blockIdx, ops := range blockOps {
		if len(ops) == 0 {
			continue
		}

		bx := blockIdx % blockCols
		by := blockIdx / blockCols

		block := spectralmath.GetBlock8(yPad, w2, bx, by)
		coeff := spectralmath.DCT8(block)

		for _, op := range ops {
			pos := midFreqPositions[op.coeffIdx]
			projected := coeff[pos.v][pos.u] * op.direction
			target := alpha * spreadTargetScale
			if projected < target {
				coeff[pos.v][pos.u] += (target - projected) * op.direction
			}
		}

		recon := spectralmath.IDCT8(coeff)
		clampBlockToByteRange(&recon)
		spectralmath.SetBlock8(yPad, w2, bx, by, recon)
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
