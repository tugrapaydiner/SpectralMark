package image

func RGBToYCbCr(img *Image) (y, cb, cr []float32) {
	if img == nil || len(img.Pix) == 0 {
		return nil, nil, nil
	}

	n := len(img.Pix)
	y = make([]float32, n)
	cb = make([]float32, n)
	cr = make([]float32, n)

	for i, p := range img.Pix {
		r := float32(p.R)
		g := float32(p.G)
		b := float32(p.B)

		y[i] = 0.299*r + 0.587*g + 0.114*b
		cb[i] = 128 - 0.168736*r - 0.331264*g + 0.5*b
		cr[i] = 128 + 0.5*r - 0.418688*g - 0.081312*b
	}

	return y, cb, cr
}

func YCbCrToRGB(w, h int, y, cb, cr []float32) *Image {
	pixelCount, _, err := checkedImageSizes(w, h)
	if err != nil {
		return &Image{W: w, H: h}
	}

	pix := make([]Rgb, pixelCount)

	for i := 0; i < pixelCount; i++ {
		yv := sampleChannel(y, i, 0)
		cbv := sampleChannel(cb, i, 128)
		crv := sampleChannel(cr, i, 128)

		r := yv + 1.402*(crv-128)
		g := yv - 0.344136*(cbv-128) - 0.714136*(crv-128)
		b := yv + 1.772*(cbv-128)

		pix[i] = Rgb{
			R: clampFloatToUint8(r),
			G: clampFloatToUint8(g),
			B: clampFloatToUint8(b),
		}
	}

	return &Image{
		W:   w,
		H:   h,
		Pix: pix,
	}
}

func sampleChannel(ch []float32, i int, fallback float32) float32 {
	if i < 0 || i >= len(ch) {
		return fallback
	}
	return ch[i]
}

func clampFloatToUint8(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(v + 0.5)
}
