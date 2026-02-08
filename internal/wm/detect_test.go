package wm

import (
	"path/filepath"
	"testing"

	spectralimage "spectralmark/internal/image"
)

func TestDetectPPMWatermarkedVsOriginal(t *testing.T) {
	dir := t.TempDir()
	origPath := filepath.Join(dir, "orig.ppm")
	wmPath := filepath.Join(dir, "wm.ppm")

	input := syntheticImage(128, 128)
	if err := spectralimage.WritePPM(origPath, input); err != nil {
		t.Fatalf("WritePPM(orig) error = %v", err)
	}
	if err := EmbedPPM(origPath, wmPath, "k", "HELLO", 3.0); err != nil {
		t.Fatalf("EmbedPPM() error = %v", err)
	}

	scoreOrig, presentOrig, msgOrig, okOrig, err := DetectPPM(origPath, "k")
	if err != nil {
		t.Fatalf("DetectPPM(orig) error = %v", err)
	}
	if presentOrig || okOrig {
		t.Fatalf("expected original image to be not present/ok, got present=%v ok=%v score=%f msg=%q", presentOrig, okOrig, scoreOrig, msgOrig)
	}

	scoreWM, presentWM, msgWM, okWM, err := DetectPPM(wmPath, "k")
	if err != nil {
		t.Fatalf("DetectPPM(wm) error = %v", err)
	}
	if !presentWM || !okWM {
		t.Fatalf("expected watermarked image to be present/ok, got present=%v ok=%v score=%f msg=%q", presentWM, okWM, scoreWM, msgWM)
	}
	if msgWM != "HELLO" {
		t.Fatalf("decoded message mismatch: got=%q want=%q", msgWM, "HELLO")
	}
	if scoreWM <= scoreOrig {
		t.Fatalf("expected watermarked score > original score, got watermarked=%f original=%f", scoreWM, scoreOrig)
	}
}

func TestDetectPPMWrongKey(t *testing.T) {
	dir := t.TempDir()
	origPath := filepath.Join(dir, "orig.ppm")
	wmPath := filepath.Join(dir, "wm.ppm")

	input := syntheticImage(128, 128)
	if err := spectralimage.WritePPM(origPath, input); err != nil {
		t.Fatalf("WritePPM(orig) error = %v", err)
	}
	if err := EmbedPPM(origPath, wmPath, "k", "HELLO", 3.0); err != nil {
		t.Fatalf("EmbedPPM() error = %v", err)
	}

	_, present, _, ok, err := DetectPPM(wmPath, "wrong-key")
	if err != nil {
		t.Fatalf("DetectPPM(wm, wrong key) error = %v", err)
	}
	if present || ok {
		t.Fatalf("expected wrong key detection to fail, got present=%v ok=%v", present, ok)
	}
}
