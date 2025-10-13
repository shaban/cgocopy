package main

import (
	"os"
	"text/template"
)

// TemplateData contains data for template rendering
type TemplateData struct {
	InputFile string
	Structs   []Struct
}

// Metadata implementation template
const metadataTemplate = `// GENERATED CODE - DO NOT EDIT
// Generated from: {{.InputFile}}

#include <stdlib.h>
#include "../../native/cgocopy_macros.h"
#include "structs.h"

{{range $struct := .Structs}}
// Metadata for {{$struct.Name}}
CGOCOPY_STRUCT({{$struct.Name}},
{{- range $idx, $field := $struct.Fields}}
    CGOCOPY_FIELD({{$struct.Name}}, {{$field.Name}}){{if ne $idx (sub1 (len $struct.Fields))}},{{end}}
{{- end}}
)

const cgocopy_struct_info* get_{{$struct.Name}}_metadata(void) {
    return &cgocopy_metadata_{{$struct.Name}};
}

{{end}}`

// API header template
const apiHeaderTemplate = `// GENERATED CODE - DO NOT EDIT
// Generated from: {{.InputFile}}

#ifndef METADATA_API_H
#define METADATA_API_H

#include "../../native/cgocopy_macros.h"

// Getter functions for each struct
{{range .Structs}}
const cgocopy_struct_info* get_{{.Name}}_metadata(void);
{{end}}

#endif // METADATA_API_H
`

// generateMetadata generates the C metadata implementation file
func generateMetadata(data TemplateData, outputFile string) error {
	tmpl := template.Must(template.New("metadata").Funcs(template.FuncMap{
		"sub1": func(n int) int { return n - 1 },
	}).Parse(metadataTemplate))

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// generateAPIHeader generates the API header file
func generateAPIHeader(data TemplateData, outputFile string) error {
	tmpl := template.Must(template.New("api").Parse(apiHeaderTemplate))

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
