package math

func PadTo8(y []float32, w, h int) (y2 []float32, w2, h2 int) {
	if w <= 0 || h <= 0 {
		return nil, 0, 0
	}

	w2 = roundUp8(w)
	h2 = roundUp8(h)
	y2 = make([]float32, w2*h2)

	for py := 0; py < h2; py++ {
		srcY := py
		if srcY >= h {
			srcY = h - 1
		}

		for px := 0; px < w2; px++ {
			srcX := px
			if srcX >= w {
				srcX = w - 1
			}

			y2[py*w2+px] = sampleAt(y, w, srcX, srcY)
		}
	}

	return y2, w2, h2
}

func Unpad(y2 []float32, w2, h2 int, w, h int) []float32 {
	if w <= 0 || h <= 0 {
		return nil
	}

	out := make([]float32, w*h)
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			out[py*w+px] = sampleAtWH(y2, w2, h2, px, py)
		}
	}

	return out
}

func GetBlock8(y []float32, w, bx, by int) [8][8]float32 {
	var b [8][8]float32
	if w <= 0 {
		return b
	}

	x0 := bx * 8
	y0 := by * 8

	for j := 0; j < 8; j++ {
		for i := 0; i < 8; i++ {
			b[j][i] = sampleAt(y, w, x0+i, y0+j)
		}
	}

	return b
}

func SetBlock8(y []float32, w, bx, by int, b [8][8]float32) {
	if w <= 0 {
		return
	}

	x0 := bx * 8
	y0 := by * 8

	for j := 0; j < 8; j++ {
		for i := 0; i < 8; i++ {
			writeAt(y, w, x0+i, y0+j, b[j][i])
		}
	}
}

func roundUp8(n int) int {
	return ((n + 7) / 8) * 8
}

func sampleAt(y []float32, w, x, yy int) float32 {
	if w <= 0 || x < 0 || yy < 0 {
		return 0
	}

	idx := yy*w + x
	if idx < 0 || idx >= len(y) {
		return 0
	}

	return y[idx]
}

func sampleAtWH(y []float32, w, h, x, yy int) float32 {
	if w <= 0 || h <= 0 || x < 0 || yy < 0 || x >= w || yy >= h {
		return 0
	}

	idx := yy*w + x
	if idx < 0 || idx >= len(y) {
		return 0
	}

	return y[idx]
}

func writeAt(y []float32, w, x, yy int, v float32) {
	if w <= 0 || x < 0 || yy < 0 {
		return
	}

	idx := yy*w + x
	if idx < 0 || idx >= len(y) {
		return
	}

	y[idx] = v
}
