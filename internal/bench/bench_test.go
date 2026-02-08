package bench

import (
	"path/filepath"
	"strings"
	"testing"

	spectralimage "spectralmark/internal/image"
)

func TestRunBenchAndFormatTable(t *testing.T) {
	origPath := filepath.Join(t.TempDir(), "orig.ppm")
	img := benchImage(128, 128)
	if err := spectralimage.WritePPM(origPath, img); err != nil {
		t.Fatalf("WritePPM(orig) error = %v", err)
	}

	results, err := RunBench(origPath, "k", "HELLO", 3.0)
	if err != nil {
		t.Fatalf("RunBench() error = %v", err)
	}
	if len(results) < 6 {
		t.Fatalf("unexpected result count: got %d want >= 6", len(results))
	}

	var base *Result
	for i := range results {
		if results[i].Attack == "none" {
			base = &results[i]
			break
		}
	}
	if base == nil {
		t.Fatalf("missing 'none' row in bench output")
	}
	if !base.Present || !base.Decode || !base.Match {
		t.Fatalf("expected 'none' attack to detect payload, got present=%v decode=%v match=%v", base.Present, base.Decode, base.Match)
	}

	table := FormatResultsTable(results)
	if !strings.Contains(table, "attack") || !strings.Contains(table, "score") {
		t.Fatalf("formatted table missing expected headers:\n%s", table)
	}
	if !strings.Contains(table, "dct-quantize") {
		t.Fatalf("formatted table missing expected attack row:\n%s", table)
	}
}

func benchImage(w, h int) *spectralimage.Image {
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
