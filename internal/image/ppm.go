package image

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

func ReadPPM(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	return readPPM(f)
}

func ReadPPMReader(r io.Reader) (*Image, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}
	return readPPM(r)
}

func readPPM(src io.Reader) (*Image, error) {
	r := bufio.NewReader(src)

	magic, err := readRequiredToken(r, "magic")
	if err != nil {
		return nil, err
	}
	if magic != "P6" {
		return nil, fmt.Errorf("unsupported magic %q (expected P6)", magic)
	}

	widthToken, err := readRequiredToken(r, "width")
	if err != nil {
		return nil, err
	}
	heightToken, err := readRequiredToken(r, "height")
	if err != nil {
		return nil, err
	}
	maxValToken, err := readRequiredToken(r, "maxval")
	if err != nil {
		return nil, err
	}

	w, err := parsePositiveInt("width", widthToken)
	if err != nil {
		return nil, err
	}
	h, err := parsePositiveInt("height", heightToken)
	if err != nil {
		return nil, err
	}

	maxVal, err := strconv.Atoi(maxValToken)
	if err != nil {
		return nil, fmt.Errorf("invalid maxval %q", maxValToken)
	}
	if maxVal != 255 {
		return nil, fmt.Errorf("unsupported maxval %d (expected 255)", maxVal)
	}

	pixelCount, byteCount, err := checkedImageSizes(w, h)
	if err != nil {
		return nil, err
	}

	raw := make([]byte, byteCount)
	if _, err := io.ReadFull(r, raw); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, fmt.Errorf("pixel data is truncated")
		}
		return nil, fmt.Errorf("read pixel data: %w", err)
	}

	pix := make([]Rgb, pixelCount)
	for i := 0; i < pixelCount; i++ {
		base := i * 3
		pix[i] = Rgb{
			R: raw[base],
			G: raw[base+1],
			B: raw[base+2],
		}
	}

	return &Image{
		W:   w,
		H:   h,
		Pix: pix,
	}, nil
}

func WritePPM(path string, img *Image) (err error) {
	if img == nil {
		return errors.New("image is nil")
	}

	pixelCount, byteCount, err := checkedImageSizes(img.W, img.H)
	if err != nil {
		return err
	}
	if len(img.Pix) != pixelCount {
		return fmt.Errorf("pixel buffer length %d does not match dimensions %dx%d", len(img.Pix), img.W, img.H)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close %s: %w", path, closeErr)
		}
	}()

	w := bufio.NewWriter(f)
	if _, err := fmt.Fprintf(w, "P6\n%d %d\n255\n", img.W, img.H); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	raw := make([]byte, byteCount)
	for i, p := range img.Pix {
		base := i * 3
		raw[base] = p.R
		raw[base+1] = p.G
		raw[base+2] = p.B
	}

	if _, err := w.Write(raw); err != nil {
		return fmt.Errorf("write pixel data: %w", err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush %s: %w", path, err)
	}

	return nil
}

func readRequiredToken(r *bufio.Reader, field string) (string, error) {
	tok, err := readHeaderToken(r)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", fmt.Errorf("missing %s in PPM header", field)
		}
		return "", fmt.Errorf("read %s in PPM header: %w", field, err)
	}
	return tok, nil
}

func readHeaderToken(r *bufio.Reader) (string, error) {
	first, err := readFirstTokenByte(r)
	if err != nil {
		return "", err
	}

	token := []byte{first}
	for {
		b, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return string(token), nil
			}
			return "", err
		}
		if isPPMWhitespace(b) {
			return string(token), nil
		}
		if b == '#' {
			if err := skipCommentLine(r); err != nil && !errors.Is(err, io.EOF) {
				return "", err
			}
			return string(token), nil
		}
		token = append(token, b)
	}
}

func readFirstTokenByte(r *bufio.Reader) (byte, error) {
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if isPPMWhitespace(b) {
			continue
		}
		if b == '#' {
			if err := skipCommentLine(r); err != nil {
				return 0, err
			}
			continue
		}
		return b, nil
	}
}

func skipCommentLine(r *bufio.Reader) error {
	for {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		if b == '\n' || b == '\r' {
			return nil
		}
	}
}

func isPPMWhitespace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
}

func parsePositiveInt(name, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", name, value)
	}
	if n <= 0 {
		return 0, fmt.Errorf("%s must be > 0", name)
	}
	return n, nil
}

func checkedImageSizes(w, h int) (pixelCount int, byteCount int, err error) {
	if w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("invalid image size %dx%d", w, h)
	}

	maxInt := int(^uint(0) >> 1)

	if w > maxInt/h {
		return 0, 0, fmt.Errorf("image dimensions overflow: %dx%d", w, h)
	}
	pixelCount = w * h

	if pixelCount > maxInt/3 {
		return 0, 0, fmt.Errorf("pixel data size overflows for %dx%d", w, h)
	}
	byteCount = pixelCount * 3

	return pixelCount, byteCount, nil
}
