package main

import (
	"flag"
	"fmt"
	"io"
	stdmath "math"
	"os"

	spectralimage "spectralmark/internal/image"
	spectralmath "spectralmark/internal/math"
	spectralwm "spectralmark/internal/wm"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return 0
	}

	switch args[0] {
	case "embed":
		return runEmbed(args[1:])
	case "prng-demo":
		return runPRNGDemo(args[1:])
	case "payload-demo":
		return runPayloadDemo(args[1:])
	case "ppm-copy":
		return runPPMCopy(args[1:])
	case "to-gray":
		return runToGray(args[1:])
	case "dct-check":
		return runDCTCheck(args[1:])
	case "detect":
		return runDetect(args[1:])
	case "bench":
		fmt.Println("TODO: bench")
		return 0
	case "serve":
		fmt.Println("TODO: serve")
		return 0
	case "metrics":
		fmt.Println("TODO: metrics")
		return 0
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage(os.Stderr)
		return 1
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: spectralmark <command>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  embed    Embed payload into PPM image")
	fmt.Fprintln(w, "  prng-demo Print deterministic PRNG samples from a key")
	fmt.Fprintln(w, "  payload-demo Encode/decode payload bits with repetition coding")
	fmt.Fprintln(w, "  ppm-copy Copy a P6 PPM file")
	fmt.Fprintln(w, "  to-gray  Convert a PPM image to grayscale")
	fmt.Fprintln(w, "  dct-check Print max reconstruction error for DCT8->IDCT8")
	fmt.Fprintln(w, "  detect   Detect watermark and recover message")
	fmt.Fprintln(w, "  bench    TODO")
	fmt.Fprintln(w, "  serve    TODO")
	fmt.Fprintln(w, "  metrics  TODO")
	fmt.Fprintln(w, "  help     Show this help")
}

func runPPMCopy(args []string) int {
	fs := flag.NewFlagSet("ppm-copy", flag.ContinueOnError)

	var inPath string
	var outPath string

	fs.StringVar(&inPath, "in", "", "input PPM path")
	fs.StringVar(&outPath, "out", "", "output PPM path")
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
		printPPMCopyUsage(os.Stderr)
		return 1
	}

	if inPath == "" || outPath == "" {
		fmt.Fprintln(os.Stderr, "both --in and --out are required")
		printPPMCopyUsage(os.Stderr)
		return 1
	}

	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		printPPMCopyUsage(os.Stderr)
		return 1
	}

	img, err := spectralimage.ReadPPM(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input PPM: %v\n", err)
		return 1
	}

	if err := spectralimage.WritePPM(outPath, img); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write output PPM: %v\n", err)
		return 1
	}

	return 0
}

func runEmbed(args []string) int {
	fs := flag.NewFlagSet("embed", flag.ContinueOnError)

	var inPath string
	var outPath string
	var key string
	var msg string
	var alpha float64

	fs.StringVar(&inPath, "in", "", "input PPM path")
	fs.StringVar(&outPath, "out", "", "output PPM path")
	fs.StringVar(&key, "key", "", "embedding key")
	fs.StringVar(&msg, "msg", "", "message payload")
	fs.Float64Var(&alpha, "alpha", 3.0, "embedding strength")
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
		printEmbedUsage(os.Stderr)
		return 1
	}

	if inPath == "" || outPath == "" || key == "" || msg == "" {
		fmt.Fprintln(os.Stderr, "--in, --out, --key, and --msg are required")
		printEmbedUsage(os.Stderr)
		return 1
	}
	if alpha <= 0 {
		fmt.Fprintln(os.Stderr, "--alpha must be > 0")
		printEmbedUsage(os.Stderr)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		printEmbedUsage(os.Stderr)
		return 1
	}

	if err := spectralwm.EmbedPPM(inPath, outPath, key, msg, float32(alpha)); err != nil {
		fmt.Fprintf(os.Stderr, "embed failed: %v\n", err)
		return 1
	}

	return 0
}

func printEmbedUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: spectralmark embed --in <input.ppm> --out <output.ppm> --key <key> --msg <msg> --alpha <strength>")
}

func printPPMCopyUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: spectralmark ppm-copy --in <input.ppm> --out <output.ppm>")
}

func runToGray(args []string) int {
	fs := flag.NewFlagSet("to-gray", flag.ContinueOnError)

	var inPath string
	var outPath string

	fs.StringVar(&inPath, "in", "", "input PPM path")
	fs.StringVar(&outPath, "out", "", "output PPM path")
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
		printToGrayUsage(os.Stderr)
		return 1
	}

	if inPath == "" || outPath == "" {
		fmt.Fprintln(os.Stderr, "both --in and --out are required")
		printToGrayUsage(os.Stderr)
		return 1
	}

	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		printToGrayUsage(os.Stderr)
		return 1
	}

	img, err := spectralimage.ReadPPM(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input PPM: %v\n", err)
		return 1
	}

	y, cb, cr := spectralimage.RGBToYCbCr(img)
	for i := range cb {
		cb[i] = 128
	}
	for i := range cr {
		cr[i] = 128
	}

	gray := spectralimage.YCbCrToRGB(img.W, img.H, y, cb, cr)
	if err := spectralimage.WritePPM(outPath, gray); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write output PPM: %v\n", err)
		return 1
	}

	return 0
}

func printToGrayUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: spectralmark to-gray --in <input.ppm> --out <output.ppm>")
}

func runDCTCheck(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(os.Stderr, "Usage: spectralmark dct-check")
		return 1
	}

	var block [8][8]float32
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			block[y][x] = float32((y*8 + x) - 32)
		}
	}

	coeff := spectralmath.DCT8(block)
	recon := spectralmath.IDCT8(coeff)

	maxErr := float32(0)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			err := float32(stdmath.Abs(float64(recon[y][x] - block[y][x])))
			if err > maxErr {
				maxErr = err
			}
		}
	}

	fmt.Printf("max reconstruction error: %.9f\n", maxErr)
	return 0
}

func runDetect(args []string) int {
	fs := flag.NewFlagSet("detect", flag.ContinueOnError)

	var inPath string
	var key string
	fs.StringVar(&inPath, "in", "", "input PPM path")
	fs.StringVar(&key, "key", "", "detection key")
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
		printDetectUsage(os.Stderr)
		return 1
	}
	if inPath == "" || key == "" {
		fmt.Fprintln(os.Stderr, "--in and --key are required")
		printDetectUsage(os.Stderr)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		printDetectUsage(os.Stderr)
		return 1
	}

	score, present, msg, ok, err := spectralwm.DetectPPM(inPath, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "detect failed: %v\n", err)
		return 1
	}

	fmt.Printf("score: %.4f\n", score)
	fmt.Printf("present: %v\n", present)
	fmt.Printf("decode ok: %v\n", ok)
	if ok {
		fmt.Printf("msg: %s\n", msg)
	}

	return 0
}

func printDetectUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: spectralmark detect --in <input.ppm> --key <key>")
}

func runPRNGDemo(args []string) int {
	fs := flag.NewFlagSet("prng-demo", flag.ContinueOnError)

	var key string
	var n int
	fs.StringVar(&key, "key", "", "key string")
	fs.IntVar(&n, "n", 10, "number of samples")
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
		printPRNGDemoUsage(os.Stderr)
		return 1
	}

	if key == "" {
		fmt.Fprintln(os.Stderr, "--key is required")
		printPRNGDemoUsage(os.Stderr)
		return 1
	}
	if n <= 0 {
		fmt.Fprintln(os.Stderr, "--n must be > 0")
		printPRNGDemoUsage(os.Stderr)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		printPRNGDemoUsage(os.Stderr)
		return 1
	}

	seed := spectralwm.SeedFromKey(key)
	rng := spectralwm.NewPRNG(seed)

	fmt.Printf("seed: %d\n", seed)
	for i := 0; i < n; i++ {
		fmt.Printf("%d\n", rng.NextU64())
	}

	return 0
}

func printPRNGDemoUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: spectralmark prng-demo --key <key> --n <count>")
}

func runPayloadDemo(args []string) int {
	fs := flag.NewFlagSet("payload-demo", flag.ContinueOnError)

	var msg string
	fs.StringVar(&msg, "msg", "HELLO", "message to encode")
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
		printPayloadDemoUsage(os.Stderr)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		printPayloadDemoUsage(os.Stderr)
		return 1
	}

	bits := spectralwm.EncodePayload(msg)
	decoded, ok := spectralwm.DecodePayload(bits)

	fmt.Printf("encoded symbols: %d\n", len(bits))
	fmt.Printf("decode ok: %v\n", ok)
	if ok {
		fmt.Printf("decoded msg: %s\n", decoded)
		return 0
	}

	return 1
}

func printPayloadDemoUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: spectralmark payload-demo --msg <message>")
}
