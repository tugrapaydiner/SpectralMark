package bench

import (
	stdmath "math"

	spectralimage "spectralmark/internal/image"
	spectralmath "spectralmark/internal/math"
	spectralwm "spectralmark/internal/wm"
)

func AttackNoise(img *spectralimage.Image, key string, sigma float32) *spectralimage.Image {
	if img == nil {
		return nil
	}
	if sigma <= 0 {
		sigma = 6
	}

	out := cloneImage(img)
	rng := spectralwm.NewPRNG(spectralwm.SeedFromKey("bench-noise:" + key))

	for i := range out.Pix {
		nr := gaussianish(rng, sigma)
		ng := gaussianish(rng, sigma)
		nb := gaussianish(rng, sigma)

		out.Pix[i] = spectralimage.Rgb{
			R: clampToByte(float32(out.Pix[i].R) + nr),
			G: clampToByte(float32(out.Pix[i].G) + ng),
			B: clampToByte(float32(out.Pix[i].B) + nb),
		}
	}

	return out
}

func AttackBrightnessContrast(img *spectralimage.Image, brightness, contrast float32) *spectralimage.Image {
	if img == nil {
		return nil
	}
	if contrast <= 0 {
		contrast = 1
	}

	out := cloneImage(img)
	for i, p := range out.Pix {
		out.Pix[i] = spectralimage.Rgb{
			R: clampToByte((float32(p.R)-128)*contrast + 128 + brightness),
			G: clampToByte((float32(p.G)-128)*contrast + 128 + brightness),
			B: clampToByte((float32(p.B)-128)*contrast + 128 + brightness),
		}
	}

	return out
}

func AttackCropCenter(img *spectralimage.Image, keepFraction float32) *spectralimage.Image {
	if img == nil {
		return nil
	}
	if keepFraction <= 0 || keepFraction > 1 {
		keepFraction = 0.85
	}

	cw := int(float32(img.W)*keepFraction + 0.5)
	ch := int(float32(img.H)*keepFraction + 0.5)
	if cw < 1 {
		cw = 1
	}
	if ch < 1 {
		ch = 1
	}

	x0 := (img.W - cw) / 2
	y0 := (img.H - ch) / 2

	crop := &spectralimage.Image{
		W:   cw,
		H:   ch,
		Pix: make([]spectralimage.Rgb, cw*ch),
	}
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			srcIdx := (y0+y)*img.W + (x0 + x)
			dstIdx := y*cw + x
			crop.Pix[dstIdx] = img.Pix[srcIdx]
		}
	}

	return ResizeNN(crop, img.W, img.H)
}

func AttackResizeNN(img *spectralimage.Image, scale float32) *spectralimage.Image {
	if img == nil {
		return nil
	}
	if scale <= 0 || scale >= 1 {
		scale = 0.75
	}

	dw := int(float32(img.W)*scale + 0.5)
	dh := int(float32(img.H)*scale + 0.5)
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}

	down := ResizeNN(img, dw, dh)
	return ResizeNN(down, img.W, img.H)
}

func AttackDCTQuantize(img *spectralimage.Image, step float32) *spectralimage.Image {
	if img == nil {
		return nil
	}
	if step <= 0 {
		step = 12
	}

	y, cb, cr := spectralimage.RGBToYCbCr(img)
	yQ := quantizeChannelDCT(y, img.W, img.H, step)
	cbQ := quantizeChannelDCT(cb, img.W, img.H, step*1.25)
	crQ := quantizeChannelDCT(cr, img.W, img.H, step*1.25)

	return spectralimage.YCbCrToRGB(img.W, img.H, yQ, cbQ, crQ)
}

func ResizeNN(img *spectralimage.Image, w, h int) *spectralimage.Image {
	if img == nil {
		return nil
	}
	if w <= 0 || h <= 0 {
		return &spectralimage.Image{}
	}

	out := &spectralimage.Image{
		W:   w,
		H:   h,
		Pix: make([]spectralimage.Rgb, w*h),
	}

	scaleX := float64(img.W) / float64(w)
	scaleY := float64(img.H) / float64(h)

	for y := 0; y < h; y++ {
		srcY := int(float64(y) * scaleY)
		if srcY >= img.H {
			srcY = img.H - 1
		}
		for x := 0; x < w; x++ {
			srcX := int(float64(x) * scaleX)
			if srcX >= img.W {
				srcX = img.W - 1
			}

			out.Pix[y*w+x] = img.Pix[srcY*img.W+srcX]
		}
	}

	return out
}

func quantizeChannelDCT(ch []float32, w, h int, step float32) []float32 {
	if w <= 0 || h <= 0 || len(ch) == 0 {
		return nil
	}
	if step <= 0 {
		step = 12
	}

	pad, w2, h2 := spectralmath.PadTo8(ch, w, h)
	if w2 == 0 || h2 == 0 {
		return nil
	}

	blocksX := w2 / 8
	blocksY := h2 / 8
	for by := 0; by < blocksY; by++ {
		for bx := 0; bx < blocksX; bx++ {
			b := spectralmath.GetBlock8(pad, w2, bx, by)
			c := spectralmath.DCT8(b)

			for v := 0; v < 8; v++ {
				for u := 0; u < 8; u++ {
					c[v][u] = float32(stdmath.Round(float64(c[v][u]/step))) * step
				}
			}

			r := spectralmath.IDCT8(c)
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					if r[y][x] < 0 {
						r[y][x] = 0
					} else if r[y][x] > 255 {
						r[y][x] = 255
					}
				}
			}

			spectralmath.SetBlock8(pad, w2, bx, by, r)
		}
	}

	return spectralmath.Unpad(pad, w2, h2, w, h)
}

func gaussianish(rng *spectralwm.PRNG, sigma float32) float32 {
	// Sum of uniforms approximates a normal distribution.
	sum := float32(0)
	for i := 0; i < 6; i++ {
		sum += rng.NextF32()
	}
	return (sum - 3.0) * sigma * 1.41421356
}

func cloneImage(img *spectralimage.Image) *spectralimage.Image {
	if img == nil {
		return nil
	}
	out := &spectralimage.Image{
		W:   img.W,
		H:   img.H,
		Pix: make([]spectralimage.Rgb, len(img.Pix)),
	}
	copy(out.Pix, img.Pix)
	return out
}

func clampToByte(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(v + 0.5)
}
