package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	input := flag.String("input", "", "Input C header file")
	output := flag.String("output", "", "Output C file (default: input_meta.c)")
	apiHeader := flag.String("api", "", "Output API header file (optional)")
	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "Usage: cgocopy-generate -input=file.h [-output=file_meta.c] [-api=api.h]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintln(os.Stderr, "  -input string")
		fmt.Fprintln(os.Stderr, "        Input C header file with struct definitions (required)")
		fmt.Fprintln(os.Stderr, "  -output string")
		fmt.Fprintln(os.Stderr, "        Output C file with metadata (default: {input}_meta.c)")
		fmt.Fprintln(os.Stderr, "  -api string")
		fmt.Fprintln(os.Stderr, "        Output API header with getter declarations (optional)")
		os.Exit(1)
	}

	// Read input file
	content, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", *input, err)
		os.Exit(1)
	}

	// Parse structs
	structs, err := parseStructs(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing structs: %v\n", err)
		os.Exit(1)
	}

	if len(structs) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: No structs found in %s\n", *input)
		os.Exit(0)
	}

	// Print found structs
	fmt.Printf("Found %d struct(s): ", len(structs))
	for i, s := range structs {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(s.Name)
	}
	fmt.Println()

	data := TemplateData{
		InputFile: *input,
		Structs:   structs,
	}

	// Generate metadata implementation
	if *output == "" {
		*output = strings.TrimSuffix(*input, ".h") + "_meta.c"
	}

	if err := generateMetadata(data, *output); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating metadata %s: %v\n", *output, err)
		os.Exit(1)
	}
	fmt.Printf("Generated: %s\n", *output)

	// Generate API header if requested
	if *apiHeader != "" {
		if err := generateAPIHeader(data, *apiHeader); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating API header: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated: %s\n", *apiHeader)
	}

	fmt.Println("✓ Code generation complete")
}
