package bench

import (
	"fmt"
	stdmath "math"
	"os"
	"path/filepath"
	"strings"

	spectralimage "spectralmark/internal/image"
	spectralutil "spectralmark/internal/util"
	spectralwm "spectralmark/internal/wm"
)

type Result struct {
	Attack  string
	PSNR    float32
	Score   float32
	Present bool
	Decode  bool
	Match   bool
	Msg     string
	Error   string
}

type attackCase struct {
	name  string
	apply func(img *spectralimage.Image) *spectralimage.Image
}

func RunBench(inPath, key, msg string, alpha float32) ([]Result, error) {
	if inPath == "" {
		return nil, fmt.Errorf("input path is required")
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

	tmpDir, err := os.MkdirTemp("", "spectralmark-bench-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	wmPath := filepath.Join(tmpDir, "watermarked.ppm")
	if err := spectralwm.EmbedPPM(inPath, wmPath, key, msg, alpha); err != nil {
		return nil, fmt.Errorf("embed input image: %w", err)
	}

	wmImg, err := spectralimage.ReadPPM(wmPath)
	if err != nil {
		return nil, fmt.Errorf("read watermarked image: %w", err)
	}
	yWM, _, _ := spectralimage.RGBToYCbCr(wmImg)

	attacks := []attackCase{
		{name: "none", apply: func(img *spectralimage.Image) *spectralimage.Image { return cloneImage(img) }},
		{name: "noise", apply: func(img *spectralimage.Image) *spectralimage.Image { return AttackNoise(img, key, 3) }},
		{name: "bright-contrast", apply: func(img *spectralimage.Image) *spectralimage.Image { return AttackBrightnessContrast(img, 5, 1.03) }},
		{name: "crop-center", apply: func(img *spectralimage.Image) *spectralimage.Image { return AttackCropCenter(img, 0.92) }},
		{name: "resize-nn", apply: func(img *spectralimage.Image) *spectralimage.Image { return AttackResizeNN(img, 0.9) }},
		{name: "dct-quantize", apply: func(img *spectralimage.Image) *spectralimage.Image { return AttackDCTQuantize(img, 10) }},
	}

	results := make([]Result, 0, len(attacks))
	for i, a := range attacks {
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

		outPath := filepath.Join(tmpDir, fmt.Sprintf("%02d_%s.ppm", i, sanitizeName(a.name)))
		if err := spectralimage.WritePPM(outPath, out); err != nil {
			row.Error = fmt.Sprintf("write attacked image: %v", err)
			results = append(results, row)
			continue
		}

		score, present, detMsg, ok, err := spectralwm.DetectPPM(outPath, key)
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

func FormatResultsTable(results []Result) string {
	if len(results) == 0 {
		return "no results\n"
	}

	attackW := len("attack")
	for _, r := range results {
		if len(r.Attack) > attackW {
			attackW = len(r.Attack)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-*s  %-10s  %-7s  %-7s  %-7s  %-7s  %s\n",
		attackW, "attack", "psnr(db)", "score", "present", "decode", "match", "msg/error")

	sepLen := attackW + 2 + 10 + 2 + 7 + 2 + 7 + 2 + 7 + 2 + 7 + 2 + 8
	b.WriteString(strings.Repeat("-", sepLen))
	b.WriteByte('\n')

	for _, r := range results {
		psnr := "n/a"
		if !stdmath.IsNaN(float64(r.PSNR)) {
			if stdmath.IsInf(float64(r.PSNR), 1) {
				psnr = "+Inf"
			} else {
				psnr = fmt.Sprintf("%.3f", r.PSNR)
			}
		}

		msg := r.Msg
		if r.Error != "" {
			msg = "ERR: " + r.Error
		}
		fmt.Fprintf(&b, "%-*s  %-10s  %7.4f  %-7t  %-7t  %-7t  %s\n",
			attackW, r.Attack, psnr, r.Score, r.Present, r.Decode, r.Match, msg)
	}

	return b.String()
}

func sanitizeName(s string) string {
	if s == "" {
		return "attack"
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b.WriteByte(c)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}
