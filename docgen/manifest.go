package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// parseManifest reads the .xojo_project file. It returns:
//   - config: the key=value settings (Type=, RBProjectVersion=, DefaultWindow=, etc.)
//   - items:  the ItemType=Name;Path;ID;ParentID;Visible declarations
func parseManifest(path string) (config map[string]string, items []ManifestItem, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	config = map[string]string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024) // long lines possible
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		it, ok := parseItemLine(line)
		if ok {
			items = append(items, it)
			continue
		}
		// key=value config line
		if eq := strings.IndexByte(line, '='); eq > 0 {
			key := strings.TrimSpace(line[:eq])
			val := strings.TrimSpace(line[eq+1:])
			// Some config values themselves contain ';' (e.g. AppIcon=Long Pepper.xojo_resources;&h0).
			// They are NOT item lines because the part before '=' is a known config key with no ';'.
			config[key] = val
		}
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	return config, items, nil
}

// parseItemLine tries to parse a line as an ItemType=Name;Path;ID;ParentID;Visible record.
// Returns ok=false if the line is not in that shape.
func parseItemLine(line string) (it ManifestItem, ok bool) {
	eq := strings.IndexByte(line, '=')
	if eq <= 0 {
		return it, false
	}
	itemType := line[:eq]
	rest := line[eq+1:]
	// Item-type prefix is alphabetic (Class, WebSession, DesktopWindow, iOSLayout, etc.).
	// Config keys like Type, RBProjectVersion, DefaultWindow would be caught here too,
	// so we validate by field shape: rest must split into exactly 5 ';'-fields and the
	// ID/ParentID fields must look like &h...
	fields := strings.Split(rest, ";")
	if len(fields) != 5 {
		return it, false
	}
	for _, f := range fields[:4] {
		// Name and Path can be non-empty; ID/ParentID must be &h...
		_ = f
	}
	idStr := strings.TrimSpace(fields[2])
	parentStr := strings.TrimSpace(fields[3])
	if !strings.HasPrefix(idStr, "&h") || !strings.HasPrefix(parentStr, "&h") {
		return it, false
	}
	id, err1 := parseHexID(idStr)
	parent, err2 := parseHexID(parentStr)
	if err1 != nil || err2 != nil {
		return it, false
	}
	// Confirm the ItemType is one we recognize as a project item (not a config key).
	if !isKnownItemType(itemType) {
		return it, false
	}
	visible := strings.TrimSpace(fields[4]) == "true"
	return ManifestItem{
		ItemType: itemType,
		Name:     strings.TrimSpace(fields[0]),
		Path:     strings.TrimSpace(fields[1]),
		ID:       id,
		ParentID: parent,
		Visible:  visible,
	}, true
}

// parseHexID parses "&h000000002DDB8FFF" or "&h0" as a uint64.
func parseHexID(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "&h")
	s = strings.TrimPrefix(s, "&H")
	return strconv.ParseUint(s, 16, 64)
}

// isKnownItemType reports whether a prefix is a real project-item type (vs a config key).
var knownItemTypes = map[string]bool{
	"Class":               true,
	"Module":              true,
	"Interface":           true,
	"BuildSteps":          true,
	"Folder":              true,
	"Library":             true,
	"MenuBar":             true,
	"DesktopWindow":       true,
	"DesktopToolbar":      true,
	"Window":              true,
	"WebSession":          true,
	"WebView":             true,
	"MobileScreen":        true,
	"MobileContainer":     true,
	"iOSLayout":           true,
	"iOSContainerControl": true,
	"MultiImage":          true,
	"AppIcons":            true,
	"ColorAsset":          true,
}

func isKnownItemType(s string) bool {
	return knownItemTypes[s]
}

// kindFor maps an ItemType prefix to a ContainerKind.
func kindFor(itemType string) ContainerKind {
	switch itemType {
	case "Class":
		return KindClass
	case "Module":
		return KindModule
	case "Interface":
		return KindInterface
	case "WebSession":
		return KindWebSession
	case "WebView", "Window", "DesktopWindow", "MobileScreen", "MobileContainer",
		"iOSLayout", "iOSContainerControl":
		return KindPage
	case "MenuBar":
		return KindMenuBar
	case "DesktopToolbar", "Toolbar":
		return KindToolbar
	case "Folder":
		return KindFolder
	case "Library":
		return KindLibrary
	case "BuildSteps":
		return KindBuildSteps
	}
	return KindOther
}

// shouldDocument reports whether a container kind is worth rendering docs for.
func shouldDocument(k ContainerKind) bool {
	switch k {
	case KindClass, KindModule, KindInterface, KindWebSession, KindPage, KindMenuBar, KindToolbar:
		return true
	}
	return false
}

// projectSlug derives a filesystem-safe slug from a project file path.
func projectSlug(projectPath string) string {
	base := filepath.Base(projectPath)
	base = strings.TrimSuffix(base, ".xojo_project")
	return slugify(base)
}

// slugify makes a filesystem- and URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	out := b.String()
	// collapse runs of '-'
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	out = strings.Trim(out, "-")
	if out == "" {
		out = "project"
	}
	return out
}

// projectDisplayName returns a nice human name for the project (title case).
func projectDisplayName(slugOrName string) string {
	// Use the .xojo_project base name, stripped of extension, with spaces preserved.
	name := filepath.Base(slugOrName)
	name = strings.TrimSuffix(name, ".xojo_project")
	return name
}

// resolveItemPath returns the absolute path to an item's source file.
// Item paths in the manifest are relative to the project directory and use '/'.
func resolveItemPath(projectDir, itemPath string) string {
	if itemPath == "" {
		return ""
	}
	// Manifest paths use forward slashes; convert for the platform.
	p := filepath.FromSlash(itemPath)
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(projectDir, p)
}

// helper for errors
func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "warn: "+format+"\n", args...)
}
