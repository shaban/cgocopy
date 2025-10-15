package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	input := flag.String("input", "", "Input C header file")
	output := flag.String("output", "", "Output C file (default: input_meta.c)")
	apiHeader := flag.String("api", "", "Output API header file (optional)")
	headerPath := flag.String("header-path", "", "Path to cgocopy_macros.h (default: auto-detect)")
	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "Usage: cgocopy-generate -input=file.h [-output=file_meta.c] [-api=api.h] [-header-path=path]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintln(os.Stderr, "  -input string")
		fmt.Fprintln(os.Stderr, "        Input C header file with struct definitions (required)")
		fmt.Fprintln(os.Stderr, "  -output string")
		fmt.Fprintln(os.Stderr, "        Output C file with metadata (default: {input}_meta.c)")
		fmt.Fprintln(os.Stderr, "  -api string")
		fmt.Fprintln(os.Stderr, "        Output API header with getter declarations (optional)")
		fmt.Fprintln(os.Stderr, "  -header-path string")
		fmt.Fprintln(os.Stderr, "        Path to cgocopy_macros.h (default: auto-detect from go.mod)")
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

	// Determine header path
	macrosPath := *headerPath
	if macrosPath == "" {
		var err error
		macrosPath, err = inferHeaderPath(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not auto-detect header path: %v\n", err)
			fmt.Fprintln(os.Stderr, "Using default: ../../native/cgocopy_macros.h")
			macrosPath = "../../native/cgocopy_macros.h"
		}
	}

	data := TemplateData{
		InputFile:  *input,
		Structs:    structs,
		MacrosPath: macrosPath,
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

// inferHeaderPath finds the path to cgocopy_macros.h relative to the output file
func inferHeaderPath(outputFile string) (string, error) {
	// Get absolute path of output file
	absOutput, err := filepath.Abs(outputFile)
	if err != nil {
		return "", err
	}

	outputDir := filepath.Dir(absOutput)

	// Walk up to find go.mod (module root)
	moduleRoot := findModuleRoot(outputDir)
	if moduleRoot == "" {
		return "", fmt.Errorf("could not find go.mod")
	}

	// Expected location of cgocopy_macros.h
	macrosPath := filepath.Join(moduleRoot, "pkg", "cgocopy", "native", "cgocopy_macros.h")

	// Check if it exists
	if _, err := os.Stat(macrosPath); err != nil {
		return "", fmt.Errorf("cgocopy_macros.h not found at %s", macrosPath)
	}

	// Calculate relative path from output directory to macros file
	relPath, err := filepath.Rel(outputDir, macrosPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

// findModuleRoot walks up from dir to find go.mod
func findModuleRoot(dir string) string {
	currentDir := dir
	for {
		modPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			return currentDir
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached root
			break
		}
		currentDir = parent
	}
	return ""
}
