package store

import (
	"path/filepath"
	"strings"
	"text/template"
)

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
