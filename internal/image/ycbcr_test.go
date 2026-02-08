package image

import "testing"

func TestRGBToYCbCrRoundTrip(t *testing.T) {
	original := &Image{
		W: 3,
		H: 1,
		Pix: []Rgb{
			{R: 255, G: 0, B: 0},
			{R: 64, G: 128, B: 200},
			{R: 5, G: 250, B: 120},
		},
	}

	y, cb, cr := RGBToYCbCr(original)
	got := YCbCrToRGB(original.W, original.H, y, cb, cr)

	if len(got.Pix) != len(original.Pix) {
		t.Fatalf("pixel count mismatch: got=%d want=%d", len(got.Pix), len(original.Pix))
	}

	for i := range original.Pix {
		assertChannelClose(t, "R", got.Pix[i].R, original.Pix[i].R, 1)
		assertChannelClose(t, "G", got.Pix[i].G, original.Pix[i].G, 1)
		assertChannelClose(t, "B", got.Pix[i].B, original.Pix[i].B, 1)
	}
}

func TestYCbCrToRGBNeutralChromaIsGrayscale(t *testing.T) {
	y := []float32{0, 42.2, 128.8, 255}
	cb := []float32{128, 128, 128, 128}
	cr := []float32{128, 128, 128, 128}

	img := YCbCrToRGB(4, 1, y, cb, cr)
	for i, p := range img.Pix {
		if p.R != p.G || p.G != p.B {
			t.Fatalf("pixel %d is not grayscale: %+v", i, p)
		}
	}
}

func assertChannelClose(t *testing.T, channel string, got, want uint8, maxDiff int) {
	t.Helper()

	diff := int(got) - int(want)
	if diff < 0 {
		diff = -diff
	}
	if diff > maxDiff {
		t.Fatalf("%s channel mismatch: got=%d want=%d diff=%d max=%d", channel, got, want, diff, maxDiff)
	}
}
