package app

import (
	"bytes"
	"encoding/json"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	spectralimage "spectralmark/internal/image"
	spectralwm "spectralmark/internal/wm"
)

func TestIndexServesHTML(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rr.Body.String(), "SpectralMark Local App") {
		t.Fatalf("response body missing expected title")
	}
}

func TestEmbedEndpointReturnsPNG(t *testing.T) {
	h := NewHandler()
	input := testPPMBytes(t, 128, 128)

	req := multipartRequest(
		t,
		"/embed",
		map[string]string{
			"key":   "k",
			"msg":   "HELLO",
			"alpha": "3.0",
		},
		"file",
		"in.ppm",
		input,
	)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Disposition"); !strings.Contains(got, "watermarked.png") {
		t.Fatalf("Content-Disposition = %q, want watermarked.png", got)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("Content-Type = %q, want image/png", ct)
	}
	if body := rr.Body.Bytes(); len(body) < 8 || !bytes.Equal(body[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		t.Fatalf("response is not a PNG file")
	}
}

func TestEmbedEndpointAcceptsPNGInput(t *testing.T) {
	h := NewHandler()
	input := testPNGBytes(t, 128, 128)

	req := multipartRequest(
		t,
		"/embed",
		map[string]string{
			"key":   "k",
			"msg":   "HELLO",
			"alpha": "3.0",
		},
		"file",
		"in.png",
		input,
	)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("Content-Type = %q, want image/png", ct)
	}
}

func TestDetectEndpointReturnsJSON(t *testing.T) {
	dir := t.TempDir()
	origPath := filepath.Join(dir, "orig.ppm")
	wmPath := filepath.Join(dir, "wm.ppm")

	img := testImage(128, 128)
	if err := spectralimage.WritePPM(origPath, img); err != nil {
		t.Fatalf("WritePPM(orig) error = %v", err)
	}
	if err := spectralwm.EmbedPPM(origPath, wmPath, "k", "HELLO", 3.0); err != nil {
		t.Fatalf("EmbedPPM() error = %v", err)
	}
	wmBytes, err := os.ReadFile(wmPath)
	if err != nil {
		t.Fatalf("ReadFile(wm) error = %v", err)
	}

	h := NewHandler()
	req := multipartRequest(
		t,
		"/detect",
		map[string]string{"key": "k"},
		"file",
		"wm.ppm",
		wmBytes,
	)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var resp detectResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v body=%q", err, rr.Body.String())
	}
	if !resp.Present || !resp.OK || resp.Msg != "HELLO" {
		t.Fatalf("unexpected detect response: %+v", resp)
	}
}

func TestEmbedThenDetectRoundTripViaHTTP(t *testing.T) {
	h := NewHandler()
	input := testPNGBytes(t, 128, 128)

	embedReq := multipartRequest(
		t,
		"/embed",
		map[string]string{
			"key":   "k",
			"msg":   "HELLO",
			"alpha": "5.0",
		},
		"file",
		"in.png",
		input,
	)
	embedRR := httptest.NewRecorder()
	h.ServeHTTP(embedRR, embedReq)
	if embedRR.Code != http.StatusOK {
		t.Fatalf("embed status = %d, want %d body=%q", embedRR.Code, http.StatusOK, embedRR.Body.String())
	}

	detectReq := multipartRequest(
		t,
		"/detect",
		map[string]string{"key": "k"},
		"file",
		"watermarked.png",
		embedRR.Body.Bytes(),
	)
	detectRR := httptest.NewRecorder()
	h.ServeHTTP(detectRR, detectReq)
	if detectRR.Code != http.StatusOK {
		t.Fatalf("detect status = %d, want %d body=%q", detectRR.Code, http.StatusOK, detectRR.Body.String())
	}

	var resp detectResponse
	if err := json.Unmarshal(detectRR.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v body=%q", err, detectRR.Body.String())
	}
	if !resp.Present || !resp.OK || resp.Msg != "HELLO" {
		t.Fatalf("unexpected detect response after roundtrip: %+v", resp)
	}
}

func multipartRequest(t *testing.T, path string, fields map[string]string, fileField, fileName string, fileData []byte) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatalf("WriteField(%q) error = %v", k, err)
		}
	}
	part, err := writer.CreateFormFile(fileField, fileName)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(fileData); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func testPPMBytes(t *testing.T, w, h int) []byte {
	t.Helper()

	path := filepath.Join(t.TempDir(), "in.ppm")
	if err := spectralimage.WritePPM(path, testImage(w, h)); err != nil {
		t.Fatalf("WritePPM() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return data
}

func testImage(w, h int) *spectralimage.Image {
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
	return &spectralimage.Image{W: w, H: h, Pix: pix}
}

func testPNGBytes(t *testing.T, w, h int) []byte {
	t.Helper()

	img := spectralimage.ToNRGBA(testImage(w, h))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}
	return buf.Bytes()
}
