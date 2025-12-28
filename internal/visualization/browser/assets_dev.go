//go:build dev

package browser

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

func LoadAssets() (http.FileSystem, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	assetPath := filepath.Join(wd, "internal/visualization/browser/assets")

	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		assetPath = filepath.Join(wd, "assets")
	}

	return http.Dir(assetPath), nil
}

func GetTemplates() (*template.Template, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	templatePath := filepath.Join(wd, "internal/visualization/browser/templates")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatePath = filepath.Join(wd, "templates")
	}

	patterns := []string{
		filepath.Join(templatePath, "*.html"),
		filepath.Join(templatePath, "partials/*.html"),
		filepath.Join(templatePath, "components/*.html"),
	}

	var tmpl *template.Template

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			content, err := os.ReadFile(match)
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
