package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LinkMap maps Xojo entity names (PascalCase, e.g. "WebButton", "Integer",
// "For...Next") to their official documentation URLs at documentation.xojo.com.
type LinkMap struct {
	byName   map[string]string // lowercase displayname -> URL
	byMember map[string]string // lowercase "class.member" -> URL
	loaded   bool
}

// LoadLinkMap parses the Sphinx objects.inv at docsRoot. Returns an empty
// (unloaded) map if the file is missing — callers should then skip linking.
func LoadLinkMap(docsRoot string) (*LinkMap, error) {
	lm := &LinkMap{byName: map[string]string{}, byMember: map[string]string{}}
	if docsRoot == "" {
		return lm, nil
	}
	invPath := filepath.Join(docsRoot, "objects.inv")
	data, err := os.ReadFile(invPath)
	if err != nil {
		return lm, nil // missing inventory is non-fatal
	}
	// objects.inv: 4 ASCII header lines, then zlib-compressed body.
	nl := bytes.IndexByte(data, '\n')
	count := 0
	for nl >= 0 {
		count++
		if count == 4 {
			break
		}
		next := bytes.IndexByte(data[nl+1:], '\n')
		if next < 0 {
			break
		}
		nl = nl + 1 + next
	}
	if count < 4 {
		return lm, fmt.Errorf("objects.inv: malformed header")
	}
	compressed := data[nl+1:]
	zr, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return lm, fmt.Errorf("objects.inv: zlib: %w", err)
	}
	defer zr.Close()
	body, err := io.ReadAll(zr)
	if err != nil {
		return lm, fmt.Errorf("objects.inv: read: %w", err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Each line: "name domain:role priority uri [displayname...]"
		// Split into exactly 5 fields on space (displayname may contain spaces).
		parts := strings.SplitN(line, " ", 5)
		if len(parts) < 4 {
			continue
		}
		domainRole := parts[1]
		uri := parts[3]
		display := ""
		if len(parts) >= 5 {
			display = parts[4]
		}
		// Only whole-page docs (std:doc) and only api/ URIs.
		if domainRole != "std:doc" {
			continue
		}
		if !strings.HasPrefix(uri, "api/") {
			continue
		}
		if strings.HasPrefix(uri, "api/xojocloud/") {
			continue // excluded by user scope
		}
		// Resolve display name: prefer the explicit displayname; else fall back
		// to the last path segment of the name.
		name := display
		if name == "-" || name == "" {
			// name field looks like "api/web/webbutton" — use last segment.
			segs := strings.Split(parts[0], "/")
			name = segs[len(segs)-1]
		}
		// Build the absolute URL.
		url := "https://documentation.xojo.com/" + uri
		// Normalize the URI to end with .html
		if !strings.HasSuffix(url, ".html") {
			url += ".html"
		}
		// Some URIs already include .html; guard against doubling.
		url = strings.ReplaceAll(url, ".html.html", ".html")
		key := strings.ToLower(name)
		if strings.Contains(key, ".") {
			// member entry like "arrays.addrow"
			if _, ok := lm.byMember[key]; !ok {
				lm.byMember[key] = url
			}
		} else {
			if _, ok := lm.byName[key]; !ok {
				lm.byName[key] = url
			}
		}
	}
	lm.loaded = true
	return lm, nil
}

// Link returns the official-doc URL for a name, and whether one was found.
func (lm *LinkMap) Link(name string) (string, bool) {
	if lm == nil || !lm.loaded {
		return "", false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}
	// Handle qualified "Class.Member".
	if i := strings.LastIndexByte(name, '.'); i > 0 {
		if url, ok := lm.byMember[strings.ToLower(name)]; ok {
			return url, true
		}
	}
	if url, ok := lm.byName[strings.ToLower(name)]; ok {
		return url, true
	}
	return "", false
}

// Loaded reports whether the inventory was successfully loaded.
func (lm *LinkMap) Loaded() bool { return lm != nil && lm.loaded }

// Count returns the number of name entries loaded.
func (lm *LinkMap) Count() int {
	if lm == nil {
		return 0
	}
	return len(lm.byName) + len(lm.byMember)
}

// defaultDocsRoot returns the standard macOS location of the Xojo docs, or "".
func defaultDocsRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	// ~/Library/Application Support/Xojo/Xojo/<version>/Documentation
	base := filepath.Join(home, "Library", "Application Support", "Xojo", "Xojo")
	entries, err := os.ReadDir(base)
	if err != nil {
		return ""
	}
	// pick the newest version directory
	var newest string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() > newest {
			newest = e.Name()
		}
	}
	if newest == "" {
		return ""
	}
	cand := filepath.Join(base, newest, "Documentation")
	if _, err := os.Stat(filepath.Join(cand, "objects.inv")); err == nil {
		return cand
	}
	return ""
}
