package math

import (
	"reflect"
	"testing"
)

func TestPadTo8AndUnpadRoundTripWeirdSize(t *testing.T) {
	w := 7
	h := 9

	orig := make([]float32, w*h)
	for i := range orig {
		orig[i] = float32(i + 1)
	}

	padded, w2, h2 := PadTo8(orig, w, h)
	if w2 != 8 || h2 != 16 {
		t.Fatalf("unexpected padded size: got %dx%d, want 8x16", w2, h2)
	}
	if len(padded) != w2*h2 {
		t.Fatalf("unexpected padded length: got %d, want %d", len(padded), w2*h2)
	}

	for y := 0; y < h; y++ {
		want := orig[y*w+(w-1)]
		got := padded[y*w2+(w2-1)]
		if got != want {
			t.Fatalf("right edge not replicated at row %d: got %v want %v", y, got, want)
		}
	}

	for x := 0; x < w2; x++ {
		want := padded[(h-1)*w2+x]
		got := padded[(h2-1)*w2+x]
		if got != want {
			t.Fatalf("bottom edge not replicated at col %d: got %v want %v", x, got, want)
		}
	}

	unpadded := Unpad(padded, w2, h2, w, h)
	if !reflect.DeepEqual(unpadded, orig) {
		t.Fatalf("Unpad() mismatch")
	}
}

func TestGetSetBlock8NoPanicOnWeirdSize(t *testing.T) {
	w := 7
	h := 9
	buf := make([]float32, w*h)

	b := [8][8]float32{}
	for j := 0; j < 8; j++ {
		for i := 0; i < 8; i++ {
			b[j][i] = float32(j*8 + i + 1)
		}
	}

	mustNotPanic(t, func() {
		SetBlock8(buf, w, 0, 1, b)
	})
	mustNotPanic(t, func() {
		_ = GetBlock8(buf, w, 0, 1)
	})
	mustNotPanic(t, func() {
		_ = GetBlock8(buf, w, 10, 10)
	})
	mustNotPanic(t, func() {
		SetBlock8(buf, w, 10, 10, b)
	})

	for x := 0; x < w; x++ {
		got := buf[8*w+x]
		want := b[0][x]
		if got != want {
			t.Fatalf("unexpected write at x=%d: got %v want %v", x, got, want)
		}
	}

	out := GetBlock8(buf, w, 0, 1)
	for x := 0; x < w; x++ {
		if out[0][x] != b[0][x] {
			t.Fatalf("unexpected read at x=%d: got %v want %v", x, out[0][x], b[0][x])
		}
	}
	for x := w; x < 8; x++ {
		if out[0][x] != 0 {
			t.Fatalf("expected zero in out-of-bounds column x=%d, got %v", x, out[0][x])
		}
	}
	for y := 1; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if out[y][x] != 0 {
				t.Fatalf("expected zero in out-of-bounds row y=%d x=%d, got %v", y, x, out[y][x])
			}
		}
	}
}

func mustNotPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	fn()
}
