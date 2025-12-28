//go:build dev

package browser

import (
	"io"
	"testing"
)

func TestLoadAssets_Dev(t *testing.T) {
	fs, err := LoadAssets()
	if err != nil {
		t.Fatalf("LoadAssets() error = %v", err)
	}

	testCases := []struct {
		name string
		path string
	}{
		{"base css", "styles/base.css"},
		{"htmx library", "scripts/vendor/htmx.min.js"},
		{"echarts library", "scripts/vendor/echarts.min.js"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := fs.Open(tc.path)
			if err != nil {
				t.Errorf("Expected %s to be accessible from filesystem, got error: %v", tc.path, err)
				return
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				t.Errorf("Failed to read %s: %v", tc.path, err)
				return
			}

			if len(content) == 0 {
				t.Errorf("Expected %s to have content, got empty file", tc.path)
			}
		})
	}
}

func TestGetTemplates_Dev(t *testing.T) {
	templates, err := GetTemplates()
	if err != nil {
		t.Fatalf("GetTemplates() error = %v", err)
	}

	if templates == nil {
		t.Error("Expected templates to be non-nil")
	}

	requiredTemplates := []string{"index.html", "chart.html"}
	for _, name := range requiredTemplates {
		tmpl := templates.Lookup(name)
		if tmpl == nil {
			t.Errorf("Expected template %s to be loaded", name)
		}
	}
}
