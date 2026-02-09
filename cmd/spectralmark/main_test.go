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

func TestRunToGrayProducesGrayscale(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.ppm")
	outPath := filepath.Join(dir, "gray.ppm")

	inputData := append(
		[]byte("P6\n2 2\n255\n"),
		[]byte{
			255, 0, 0,
			0, 255, 0,
			0, 0, 255,
			120, 80, 40,
		}...,
	)
	if err := os.WriteFile(inPath, inputData, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	code := run([]string{"to-gray", "--in", inPath, "--out", outPath})
	if code != 0 {
		t.Fatalf("run() returned %d, want 0", code)
	}

	img, err := spectralimage.ReadPPM(outPath)
	if err != nil {
		t.Fatalf("ReadPPM() error = %v", err)
	}

	for i, p := range img.Pix {
		if p.R != p.G || p.G != p.B {
			t.Fatalf("pixel %d is not grayscale: %+v", i, p)
		}
	}
}

func TestRunDCTCheck(t *testing.T) {
	code := run([]string{"dct-check"})
	if code != 0 {
		t.Fatalf("run() returned %d, want 0", code)
	}
}

func TestRunPRNGDemo(t *testing.T) {
	code := run([]string{"prng-demo", "--key", "abc", "--n", "3"})
	if code != 0 {
		t.Fatalf("run() returned %d, want 0", code)
	}
}

func TestRunPayloadDemo(t *testing.T) {
	code := run([]string{"payload-demo", "--msg", "HELLO"})
	if code != 0 {
		t.Fatalf("run() returned %d, want 0", code)
	}
}

func TestRunEmbed(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.ppm")
	outPath := filepath.Join(dir, "out.ppm")

	input := &spectralimage.Image{
		W:   128,
		H:   128,
		Pix: make([]spectralimage.Rgb, 128*128),
	}
	for y := 0; y < input.H; y++ {
		for x := 0; x < input.W; x++ {
			i := y*input.W + x
			input.Pix[i] = spectralimage.Rgb{
				R: uint8((x + y) % 256),
				G: uint8((2*x + y) % 256),
				B: uint8((x + 3*y) % 256),
			}
		}
	}

	if err := spectralimage.WritePPM(inPath, input); err != nil {
		t.Fatalf("WritePPM() error = %v", err)
	}

	code := run([]string{
		"embed",
		"--in", inPath,
		"--out", outPath,
		"--key", "k",
		"--msg", "HELLO",
		"--alpha", "3.0",
	})
	if code != 0 {
		t.Fatalf("run() returned %d, want 0", code)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file, got stat error: %v", err)
	}
}

func TestRunDetect(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.ppm")
	wmPath := filepath.Join(dir, "wm.ppm")

	input := &spectralimage.Image{
		W:   128,
		H:   128,
		Pix: make([]spectralimage.Rgb, 128*128),
	}
	for y := 0; y < input.H; y++ {
		for x := 0; x < input.W; x++ {
			i := y*input.W + x
			input.Pix[i] = spectralimage.Rgb{
				R: uint8((x + y) % 256),
				G: uint8((2*x + y) % 256),
				B: uint8((x + 3*y) % 256),
			}
		}
	}
	if err := spectralimage.WritePPM(inPath, input); err != nil {
		t.Fatalf("WritePPM() error = %v", err)
	}

	code := run([]string{
		"embed",
		"--in", inPath,
		"--out", wmPath,
		"--key", "k",
		"--msg", "HELLO",
		"--alpha", "3.0",
	})
	if code != 0 {
		t.Fatalf("embed run() returned %d, want 0", code)
	}

	code = run([]string{
		"detect",
		"--in", wmPath,
		"--key", "k",
	})
	if code != 0 {
		t.Fatalf("detect run() returned %d, want 0", code)
	}
}

func TestRunMetrics(t *testing.T) {
	dir := t.TempDir()
	aPath := filepath.Join(dir, "a.ppm")
	bPath := filepath.Join(dir, "b.ppm")
	diffPath := filepath.Join(dir, "diff.ppm")

	imgA := &spectralimage.Image{
		W:   16,
		H:   16,
		Pix: make([]spectralimage.Rgb, 16*16),
	}
	imgB := &spectralimage.Image{
		W:   16,
		H:   16,
		Pix: make([]spectralimage.Rgb, 16*16),
	}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			i := y*16 + x
			base := uint8((x + y) % 256)
			imgA.Pix[i] = spectralimage.Rgb{R: base, G: base, B: base}
			imgB.Pix[i] = spectralimage.Rgb{R: base + 1, G: base, B: base}
		}
	}

	if err := spectralimage.WritePPM(aPath, imgA); err != nil {
		t.Fatalf("WritePPM(a) error = %v", err)
	}
	if err := spectralimage.WritePPM(bPath, imgB); err != nil {
		t.Fatalf("WritePPM(b) error = %v", err)
	}

	code := run([]string{
		"metrics",
		"--a", aPath,
		"--b", bPath,
		"--diff", diffPath,
	})
	if code != 0 {
		t.Fatalf("run() returned %d, want 0", code)
	}

	if _, err := os.Stat(diffPath); err != nil {
		t.Fatalf("expected diff image file, got stat error: %v", err)
	}
}

func TestRunBench(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.ppm")

	img := &spectralimage.Image{
		W:   128,
		H:   128,
		Pix: make([]spectralimage.Rgb, 128*128),
	}
	for y := 0; y < img.H; y++ {
		for x := 0; x < img.W; x++ {
			i := y*img.W + x
			img.Pix[i] = spectralimage.Rgb{
				R: uint8((x + y) % 256),
				G: uint8((2*x + y) % 256),
				B: uint8((x + 3*y) % 256),
			}
		}
	}
	if err := spectralimage.WritePPM(inPath, img); err != nil {
		t.Fatalf("WritePPM() error = %v", err)
	}

	code := run([]string{
		"bench",
		"--in", inPath,
		"--key", "k",
		"--msg", "HELLO",
	})
	if code != 0 {
		t.Fatalf("run() returned %d, want 0", code)
	}
}

func TestRunServeInvalidPort(t *testing.T) {
	code := run([]string{"serve", "--port", "0"})
	if code == 0 {
		t.Fatalf("run() returned %d, want non-zero", code)
	}
}
