package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	spectralimage "spectralmark/internal/image"
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
		fmt.Println("TODO: embed")
		return 0
	case "ppm-copy":
		return runPPMCopy(args[1:])
	case "to-gray":
		return runToGray(args[1:])
	case "detect":
		fmt.Println("TODO: detect")
		return 0
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
	fmt.Fprintln(w, "  embed    TODO")
	fmt.Fprintln(w, "  ppm-copy Copy a P6 PPM file")
	fmt.Fprintln(w, "  to-gray  Convert a PPM image to grayscale")
	fmt.Fprintln(w, "  detect   TODO")
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
