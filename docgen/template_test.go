package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateTemplateDirRequiresCompleteTemplate(t *testing.T) {
	templateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(templateDir, "mkdocs.base.yml"), []byte("theme: material\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := validateTemplateDir(templateDir)
	if err == nil || !strings.Contains(err.Error(), "assets/featured.png") {
		t.Fatalf("validateTemplateDir error = %v, want missing required asset", err)
	}
}

func TestValidateAndCopyTemplateDir(t *testing.T) {
	templateDir := createCompleteTestTemplate(t)
	resolved, err := validateTemplateDir(templateDir)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(resolved) {
		t.Fatalf("resolved path is not absolute: %s", resolved)
	}

	destinationDir := t.TempDir()
	staleCSS := filepath.Join(destinationDir, "stylesheets", "extra.css")
	if err := os.MkdirAll(filepath.Dir(staleCSS), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staleCSS, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyTemplateDir(resolved, destinationDir); err != nil {
		t.Fatal(err)
	}

	copied, err := os.ReadFile(staleCSS)
	if err != nil {
		t.Fatal(err)
	}
	if string(copied) != "stylesheets/extra.css" {
		t.Fatalf("copied CSS = %q", copied)
	}
}

func TestPathsOverlap(t *testing.T) {
	root := t.TempDir()
	templateDir := filepath.Join(root, "templates", "custom")
	outputDir := filepath.Join(root, "docs", "api", "project")

	if pathsOverlap(templateDir, outputDir) {
		t.Fatal("separate template and output directories reported as overlapping")
	}
	if !pathsOverlap(templateDir, filepath.Join(templateDir, "generated")) {
		t.Fatal("output nested under template was not reported as overlapping")
	}
	if !pathsOverlap(filepath.Join(outputDir, "template"), outputDir) {
		t.Fatal("template nested under output was not reported as overlapping")
	}
}

func createCompleteTestTemplate(t *testing.T) string {
	t.Helper()
	templateDir := t.TempDir()
	for _, relative := range requiredTemplateFiles {
		path := filepath.Join(templateDir, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(relative), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return templateDir
}
