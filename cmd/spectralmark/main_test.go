package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	spectralimage "spectralmark/internal/image"
)

func TestRunPPMCopyRoundTrip(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.ppm")
	outPath := filepath.Join(dir, "out.ppm")

	inputData := append(
		[]byte("P6\n# comment line\n2 1\n255\n"),
		[]byte{35, 0, 1, 2, 3, 4}...,
	)
	if err := os.WriteFile(inPath, inputData, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	code := run([]string{"ppm-copy", "--in", inPath, "--out", outPath})
	if code != 0 {
		t.Fatalf("run() returned %d, want 0", code)
	}

	inImage, err := spectralimage.ReadPPM(inPath)
	if err != nil {
		t.Fatalf("ReadPPM(in) error = %v", err)
	}
	outImage, err := spectralimage.ReadPPM(outPath)
	if err != nil {
		t.Fatalf("ReadPPM(out) error = %v", err)
	}

	if !reflect.DeepEqual(outImage, inImage) {
		t.Fatalf("image mismatch after ppm-copy:\n got: %#v\nwant: %#v", outImage, inImage)
	}
}
