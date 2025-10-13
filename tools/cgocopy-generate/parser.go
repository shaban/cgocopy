package main

import (
	"regexp"
	"strings"
)

// Field represents a C struct field
type Field struct {
	Name      string
	Type      string
	ArraySize string // empty if not array
}

// Struct represents a C struct definition
type Struct struct {
	Name   string
	Fields []Field
}

// parseStructs extracts struct definitions from C code
func parseStructs(content string) ([]Struct, error) {
	var structs []Struct

	// Remove comments first
	content = removeComments(content)

	// Match struct definitions:
	// - typedef struct { ... } Name;
	// - struct Name { ... };
	// - typedef struct Name { ... } Name;
	structRegex := regexp.MustCompile(`(?:typedef\s+)?struct(?:\s+(\w+))?\s*\{([^}]+)\}(?:\s*(\w+))?`)
	matches := structRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		// Extract struct name (either before or after body)
		name := match[1]
		if name == "" {
			name = match[3]
		}
		if name == "" {
			continue // Anonymous struct, skip
		}

		body := match[2]
		s := Struct{Name: name}

		// Parse fields: type name; or type name[size];
		// Handles: int x; char* name; double arr[10]; Point3D position;
		fieldRegex := regexp.MustCompile(`([a-zA-Z_][\w\s\*]+?)\s+(\w+)(?:\[(\d+)\])?\s*;`)
		fieldMatches := fieldRegex.FindAllStringSubmatch(body, -1)

		for _, fm := range fieldMatches {
			fieldType := strings.TrimSpace(fm[1])
			fieldName := fm[2]
			arraySize := fm[3]

			s.Fields = append(s.Fields, Field{
				Name:      fieldName,
				Type:      fieldType,
				ArraySize: arraySize,
			})
		}

		if len(s.Fields) > 0 {
			structs = append(structs, s)
		}
	}

	return structs, nil
}

// removeComments removes C/C++ style comments from source code
func removeComments(content string) string {
	// Remove // comments (line comments)
	lineComment := regexp.MustCompile(`//.*`)
	content = lineComment.ReplaceAllString(content, "")

	// Remove /* */ comments (block comments)
	blockComment := regexp.MustCompile(`(?s)/\*.*?\*/`)
	content = blockComment.ReplaceAllString(content, "")

	return content
}
