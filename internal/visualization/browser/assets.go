

//go:build !dev

package browser

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

//go:embed assets/**/* templates/**/*
var assetsFS embed.FS

func LoadAssets() (http.FileSystem, error) {
	fsys, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		return nil, err
	}
	return http.FS(fsys), nil
}

func GetTemplates() (*template.Template, error) {
	patterns := []string{
		"templates/*.html",
		"templates/partials/*.html",
		"templates/components/*.html",
	}

	var tmpl *template.Template

	for _, pattern := range patterns {
		matches, err := fs.Glob(assetsFS, pattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			content, err := fs.ReadFile(assetsFS, match)
			if err != nil {
				return nil, err
			}

			name := filepath.Base(match)

			if tmpl == nil {
				tmpl = template.New(name)
				_, err = tmpl.Parse(string(content))
			} else {
				_, err = tmpl.New(name).Parse(string(content))
			}

			if err != nil {
				return nil, err
			}
		}
	}

	if tmpl == nil {
		return template.New(""), nil
	}

	return tmpl, nil
}
