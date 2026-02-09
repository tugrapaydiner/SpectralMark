package image

import stdimage "image"

func FromStdImage(src stdimage.Image) *Image {
	if src == nil {
		return &Image{}
	}

	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return &Image{}
	}

	pix := make([]Rgb, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl, _ := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			pix[y*w+x] = Rgb{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(bl >> 8),
			}
		}
	}

	return &Image{
		W:   w,
		H:   h,
		Pix: pix,
	}
}

func ToNRGBA(src *Image) *stdimage.NRGBA {
	if src == nil || src.W <= 0 || src.H <= 0 {
		return stdimage.NewNRGBA(stdimage.Rect(0, 0, 0, 0))
	}

	out := stdimage.NewNRGBA(stdimage.Rect(0, 0, src.W, src.H))
	pixelCount := src.W * src.H
	for i := 0; i < pixelCount; i++ {
		x := i % src.W
		y := i / src.W
		offset := y*out.Stride + x*4

		var p Rgb
		if i < len(src.Pix) {
			p = src.Pix[i]
		}
		out.Pix[offset] = p.R
		out.Pix[offset+1] = p.G
		out.Pix[offset+2] = p.B
		out.Pix[offset+3] = 255
	}

	return out
}
