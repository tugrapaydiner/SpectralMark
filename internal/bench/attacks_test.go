package bench

import (
	"reflect"
	"testing"

	spectralimage "spectralmark/internal/image"
)

func TestAttacksKeepDimensions(t *testing.T) {
	img := testImage(64, 48)

	tests := []struct {
		name string
		fn   func(*spectralimage.Image) *spectralimage.Image
	}{
		{name: "noise", fn: func(in *spectralimage.Image) *spectralimage.Image { return AttackNoise(in, "k", 6) }},
		{name: "brightness-contrast", fn: func(in *spectralimage.Image) *spectralimage.Image { return AttackBrightnessContrast(in, 12, 1.1) }},
		{name: "crop-center", fn: func(in *spectralimage.Image) *spectralimage.Image { return AttackCropCenter(in, 0.85) }},
		{name: "resize-nn", fn: func(in *spectralimage.Image) *spectralimage.Image { return AttackResizeNN(in, 0.75) }},
		{name: "dct-quantize", fn: func(in *spectralimage.Image) *spectralimage.Image { return AttackDCTQuantize(in, 14) }},
	}

	for _, tt := range tests {
		out := tt.fn(img)
		if out == nil {
			t.Fatalf("%s returned nil image", tt.name)
		}
		if out.W != img.W || out.H != img.H {
			t.Fatalf("%s changed dimensions: got %dx%d want %dx%d", tt.name, out.W, out.H, img.W, img.H)
		}
		if len(out.Pix) != len(img.Pix) {
			t.Fatalf("%s changed pixel buffer length: got %d want %d", tt.name, len(out.Pix), len(img.Pix))
		}
	}
}

func TestAttackNoiseDeterministicByKey(t *testing.T) {
	img := testImage(32, 32)

	a := AttackNoise(img, "k", 6)
	b := AttackNoise(img, "k", 6)
	c := AttackNoise(img, "k2", 6)

	if !reflect.DeepEqual(a, b) {
		t.Fatalf("noise attack should be deterministic for same key")
	}
	if reflect.DeepEqual(a, c) {
		t.Fatalf("noise attack should differ for different keys")
	}
}

func testImage(w, h int) *spectralimage.Image {
	pix := make([]spectralimage.Rgb, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			pix[i] = spectralimage.Rgb{
				R: uint8((x + y) % 256),
				G: uint8((2*x + y) % 256),
				B: uint8((x + 3*y) % 256),
			}
		}
	}
	return &spectralimage.Image{
		W:   w,
		H:   h,
		Pix: pix,
	}
}
