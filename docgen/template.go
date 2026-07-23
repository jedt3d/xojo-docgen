package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var requiredTemplateFiles = []string{
	"mkdocs.base.yml",
	"assets/featured.png",
	"assets/featured_portrait.png",
	"javascripts/landing-sidebar.js",
	"javascripts/prism.js",
	"javascripts/source-modal.js",
	"javascripts/xojo.prism.js",
	"javascripts/editorial.js",
	"overrides/main.html",
	"stylesheets/editorial.css",
	"stylesheets/extra.css",
	"stylesheets/primary-color.css",
}

// resolveTemplateDir selects an explicit project template or finds the default
// template distributed beside the xojo-docgen executable.
func resolveTemplateDir(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return validateTemplateDir(explicit)
	}

	var candidates []string
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "templates", "default"))
	}
	if workingDir, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(workingDir, "templates", "default"),
			filepath.Join(workingDir, "docgen", "templates", "default"),
		)
	}

	seen := make(map[string]bool)
	for _, candidate := range candidates {
		absolute, err := filepath.Abs(candidate)
		if err != nil || seen[absolute] {
			continue
		}
		seen[absolute] = true
		info, err := os.Stat(absolute)
		if err == nil && info.IsDir() {
			return validateTemplateDir(absolute)
		}
	}

	return "", fmt.Errorf("default template directory not found; expected templates/default beside xojo-docgen or provide -template-dir")
}

func validateTemplateDir(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("%s: %w", absolute, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", absolute)
	}
	for _, relative := range requiredTemplateFiles {
		templateFile := filepath.Join(absolute, filepath.FromSlash(relative))
		fileInfo, err := os.Stat(templateFile)
		if err != nil {
			return "", fmt.Errorf("%s is incomplete: missing %s", absolute, relative)
		}
		if !fileInfo.Mode().IsRegular() {
			return "", fmt.Errorf("%s is incomplete: %s is not a regular file", absolute, relative)
		}
	}
	return absolute, nil
}

func copyTemplateDir(sourceDir string, destinationDir string) error {
	return filepath.WalkDir(sourceDir, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(sourceDir, sourcePath)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("template symlinks are not supported: %s", sourcePath)
		}

		destinationPath := filepath.Join(destinationDir, relative)
		if entry.IsDir() {
			return os.MkdirAll(destinationPath, 0o755)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("unsupported template entry: %s", sourcePath)
		}
		if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
			return err
		}
		return copyTemplateFile(sourcePath, destinationPath)
	})
}

func copyTemplateFile(sourcePath string, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	info, err := source.Stat()
	if err != nil {
		return err
	}
	destination, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(destination, source); err != nil {
		destination.Close()
		return err
	}
	return destination.Close()
}

func pathsOverlap(first string, second string) bool {
	firstAbsolute, firstErr := filepath.Abs(first)
	secondAbsolute, secondErr := filepath.Abs(second)
	if firstErr != nil || secondErr != nil {
		return true
	}
	return pathContains(firstAbsolute, secondAbsolute) || pathContains(secondAbsolute, firstAbsolute)
}

func pathContains(parent string, child string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}
