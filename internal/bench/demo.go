package bench

import (
	"fmt"
	stdmath "math"

	spectralimage "spectralmark/internal/image"
	spectralutil "spectralmark/internal/util"
	spectralwm "spectralmark/internal/wm"
)

func RunDemo(inPath, outPath, key, msg string, alpha float32) ([]Result, error) {
	if inPath == "" {
		return nil, fmt.Errorf("input path is required")
	}
	if outPath == "" {
		return nil, fmt.Errorf("output path is required")
	}
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	if msg == "" {
		return nil, fmt.Errorf("message is required")
	}
	if alpha <= 0 {
		return nil, fmt.Errorf("alpha must be > 0")
	}

	if err := spectralwm.EmbedPPM(inPath, outPath, key, msg, alpha); err != nil {
		return nil, fmt.Errorf("embed input image: %w", err)
	}

	wmImg, err := spectralimage.ReadPPM(outPath)
	if err != nil {
		return nil, fmt.Errorf("read watermarked image: %w", err)
	}
	yWM, _, _ := spectralimage.RGBToYCbCr(wmImg)

	attacks := []attackCase{
		{name: "noise", apply: func(img *spectralimage.Image) *spectralimage.Image { return AttackNoise(img, key, 1.5) }},
		{name: "resize-nn", apply: func(img *spectralimage.Image) *spectralimage.Image { return AttackResizeNN(img, 0.99) }},
		{name: "dct-quantize", apply: func(img *spectralimage.Image) *spectralimage.Image { return AttackDCTQuantize(img, 6) }},
	}

	results := make([]Result, 0, len(attacks))
	for _, a := range attacks {
		out := a.apply(wmImg)
		row := Result{
			Attack: a.name,
			PSNR:   float32(stdmath.NaN()),
		}
		if out == nil || len(out.Pix) == 0 {
			row.Error = "attack returned empty image"
			results = append(results, row)
			continue
		}

		if out.W == wmImg.W && out.H == wmImg.H {
			yOut, _, _ := spectralimage.RGBToYCbCr(out)
			row.PSNR = spectralutil.PSNR(yWM, yOut)
		}

		score, present, detMsg, ok, err := spectralwm.DetectImage(out, key)
		row.Score = score
		row.Present = present
		row.Decode = ok
		row.Msg = detMsg
		row.Match = ok && detMsg == msg
		if err != nil {
			row.Error = err.Error()
		}

		results = append(results, row)
	}

	return results, nil
}
