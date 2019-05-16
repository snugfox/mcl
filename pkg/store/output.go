package store

import (
	"path/filepath"
	"strings"
	"text/template"
)

// BaseDir returns a path for a specified store directory, structure template,
// edition, and version. The structure template is parsed as a template.Template
// with two fields: .Edition and .Version.
func BaseDir(storeDir, structureTmpl, edition, version string) (string, error) {
	tmpl, err := template.New("dirStructure").Parse(structureTmpl)
	if err != nil {
		return "", err
	}

	var dir strings.Builder
	wrapper := struct { // Wraps fields available for the template string
		Edition string
		Version string
	}{
		Edition: edition,
		Version: version,
	}
	if err := tmpl.Execute(&dir, wrapper); err != nil {
		return "", err
	}
	return filepath.Join(storeDir, dir.String()), nil
}
