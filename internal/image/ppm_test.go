package image

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadWritePPMRoundTrip(t *testing.T) {
	original := &Image{
		W: 2,
		H: 2,
		Pix: []Rgb{
			{R: 0, G: 35, B: 10},
			{R: 255, G: 32, B: 9},
			{R: 13, G: 10, B: 35},
			{R: 1, G: 2, B: 3},
		},
	}

	path := filepath.Join(t.TempDir(), "roundtrip.ppm")
	if err := WritePPM(path, original); err != nil {
		t.Fatalf("WritePPM() error = %v", err)
	}

	got, err := ReadPPM(path)
	if err != nil {
		t.Fatalf("ReadPPM() error = %v", err)
	}

	if !reflect.DeepEqual(got, original) {
		t.Fatalf("ReadPPM() mismatch:\n got: %#v\nwant: %#v", got, original)
	}
}

func TestReadPPMSkipsHeaderComments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "comments.ppm")
	data := append(
		[]byte("P6\n# format comment\n2 1\n# max comment\n255\n"),
		[]byte{9, 8, 7, 6, 5, 4}...,
	)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	img, err := ReadPPM(path)
	if err != nil {
		t.Fatalf("ReadPPM() error = %v", err)
	}

	want := &Image{
		W: 2,
		H: 1,
		Pix: []Rgb{
			{R: 9, G: 8, B: 7},
			{R: 6, G: 5, B: 4},
		},
	}
	if !reflect.DeepEqual(img, want) {
		t.Fatalf("ReadPPM() mismatch:\n got: %#v\nwant: %#v", img, want)
	}
}
