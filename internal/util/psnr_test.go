package util

import stdmath "math"
import "testing"

func TestPSNRIdenticalIsInf(t *testing.T) {
	a := []float32{0, 10, 20, 30}
	b := []float32{0, 10, 20, 30}

	got := PSNR(a, b)
	if !stdmath.IsInf(float64(got), 1) {
		t.Fatalf("expected +Inf PSNR, got %f", got)
	}
}

func TestPSNRKnownValue(t *testing.T) {
	a := []float32{0}
	b := []float32{10}

	got := PSNR(a, b)
	want := float32(28.130804)
	if abs32(got-want) > 1e-3 {
		t.Fatalf("PSNR mismatch: got=%f want=%f", got, want)
	}
}

func TestPSNRInvalidInput(t *testing.T) {
	if got := PSNR(nil, nil); got != 0 {
		t.Fatalf("expected 0 for empty input, got %f", got)
	}
	if got := PSNR([]float32{1}, []float32{1, 2}); got != 0 {
		t.Fatalf("expected 0 for mismatched input lengths, got %f", got)
	}
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
