package wm

import (
	"path/filepath"
	"strings"
	"testing"

	spectralimage "spectralmark/internal/image"
)

func TestEmbedPPMProducesNearIdenticalOutput(t *testing.T) {
	inPath := filepath.Join(t.TempDir(), "in.ppm")
	outPath := filepath.Join(t.TempDir(), "out.ppm")

	input := syntheticImage(128, 128)
	if err := spectralimage.WritePPM(inPath, input); err != nil {
		t.Fatalf("WritePPM(input) error = %v", err)
	}

	if err := EmbedPPM(inPath, outPath, "k", "HELLO", 3.0); err != nil {
		t.Fatalf("EmbedPPM() error = %v", err)
	}

	output, err := spectralimage.ReadPPM(outPath)
	if err != nil {
		t.Fatalf("ReadPPM(output) error = %v", err)
	}

	if output.W != input.W || output.H != input.H {
		t.Fatalf("size mismatch: got %dx%d want %dx%d", output.W, output.H, input.W, input.H)
	}

	maxDiff := 0
	totalDiff := 0
	diffCount := 0
	for i := range input.Pix {
		dr := absInt(int(output.Pix[i].R) - int(input.Pix[i].R))
		dg := absInt(int(output.Pix[i].G) - int(input.Pix[i].G))
		db := absInt(int(output.Pix[i].B) - int(input.Pix[i].B))

		pixDiff := dr
		if dg > pixDiff {
			pixDiff = dg
		}
		if db > pixDiff {
			pixDiff = db
		}
		if pixDiff > maxDiff {
			maxDiff = pixDiff
		}
		if pixDiff > 0 {
			diffCount++
		}
		totalDiff += dr + dg + db
	}

	if diffCount == 0 {
		t.Fatalf("expected some pixel changes after embedding")
	}

	avgAbsDiff := float64(totalDiff) / float64(len(input.Pix)*3)
	if maxDiff > 20 {
		t.Fatalf("image changed too much: maxDiff=%d", maxDiff)
	}
	if avgAbsDiff > 2.0 {
		t.Fatalf("image changed too much: avgAbsDiff=%f", avgAbsDiff)
	}
}

func TestEmbedPPMRejectsInsufficientCapacity(t *testing.T) {
	inPath := filepath.Join(t.TempDir(), "small.ppm")
	outPath := filepath.Join(t.TempDir(), "small_out.ppm")

	small := syntheticImage(8, 8)
	if err := spectralimage.WritePPM(inPath, small); err != nil {
		t.Fatalf("WritePPM(input) error = %v", err)
	}

	err := EmbedPPM(inPath, outPath, "k", "HELLO", 3.0)
	if err == nil {
		t.Fatalf("expected capacity error, got nil")
	}
	if !strings.Contains(err.Error(), "payload too large") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func syntheticImage(w, h int) *spectralimage.Image {
	pix := make([]spectralimage.Rgb, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			pix[i] = spectralimage.Rgb{
				R: uint8((3*x + y) % 256),
				G: uint8((x + 2*y) % 256),
				B: uint8((5*x + 7*y) % 256),
			}
		}
	}
	return &spectralimage.Image{
		W:   w,
		H:   h,
		Pix: pix,
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
