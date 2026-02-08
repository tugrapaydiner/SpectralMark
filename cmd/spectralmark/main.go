package main

import (
	"fmt"
	"io"
	"os"
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
	fmt.Fprintln(w, "  detect   TODO")
	fmt.Fprintln(w, "  bench    TODO")
	fmt.Fprintln(w, "  serve    TODO")
	fmt.Fprintln(w, "  metrics  TODO")
	fmt.Fprintln(w, "  help     Show this help")
}
