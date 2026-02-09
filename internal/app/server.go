package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdimage "image"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	spectralimage "spectralmark/internal/image"
	spectralwm "spectralmark/internal/wm"
)

const maxUploadBytes int64 = 64 << 20

type detectResponse struct {
	Score   float32 `json:"score"`
	Present bool    `json:"present"`
	Msg     string  `json:"msg"`
	OK      bool    `json:"ok"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/embed", handleEmbed)
	mux.HandleFunc("/detect", handleDetect)
	return mux
}

func Serve(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, NewHandler())
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, indexHTML)
}

func handleEmbed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	key := strings.TrimSpace(r.FormValue("key"))
	msg := r.FormValue("msg")
	alpha, err := parseAlpha(r.FormValue("alpha"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}
	if msg == "" {
		http.Error(w, "msg is required", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, err := decodeUploadImage(file, fileHeader.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wmImg, err := spectralwm.EmbedImage(img, key, msg, alpha)
	if err != nil {
		http.Error(w, fmt.Sprintf("embed failed: %v", err), http.StatusBadRequest)
		return
	}

	var out bytes.Buffer
	if err := png.Encode(&out, spectralimage.ToNRGBA(wmImg)); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode output image: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="watermarked.png"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out.Bytes())
}

func handleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse multipart form: %v", err))
		return
	}

	key := strings.TrimSpace(r.FormValue("key"))
	if key == "" {
		writeJSONError(w, http.StatusBadRequest, "key is required")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to read uploaded file: %v", err))
		return
	}
	defer file.Close()

	img, err := decodeUploadImage(file, fileHeader.Filename)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	score, present, msg, ok, err := spectralwm.DetectImage(img, key)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("detect failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, detectResponse{
		Score:   score,
		Present: present,
		Msg:     msg,
		OK:      ok,
	})
}

func parseAlpha(raw string) (float32, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 5.0, nil
	}

	v, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid alpha: %q", raw)
	}
	if v <= 0 {
		return 0, fmt.Errorf("alpha must be > 0")
	}
	return float32(v), nil
}

func decodeUploadImage(file io.Reader, filename string) (*spectralimage.Image, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file body: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("uploaded file is empty")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".ppm" {
		img, err := spectralimage.ReadPPMReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to read PPM: %w", err)
		}
		return img, nil
	}

	stdImg, _, err := stdimage.Decode(bytes.NewReader(data))
	if err == nil {
		return spectralimage.FromStdImage(stdImg), nil
	}

	ppmImg, ppmErr := spectralimage.ReadPPMReader(bytes.NewReader(data))
	if ppmErr == nil {
		return ppmImg, nil
	}

	return nil, fmt.Errorf("unsupported image format; use .ppm, .png, .jpg, or .jpeg")
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
