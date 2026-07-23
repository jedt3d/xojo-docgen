package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSiteContentVersionTracksGeneratedContent(t *testing.T) {
	sourceDir := t.TempDir()
	assetPath := filepath.Join(sourceDir, "stylesheets", "editorial.css")
	if err := os.MkdirAll(filepath.Dir(assetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assetPath, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, err := siteContentVersion(sourceDir)
	if err != nil {
		t.Fatal(err)
	}
	second, err := siteContentVersion(sourceDir)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("unchanged content versions differ: %q != %q", first, second)
	}

	if err := os.WriteFile(assetPath, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := siteContentVersion(sourceDir)
	if err != nil {
		t.Fatal(err)
	}
	if changed == first {
		t.Fatalf("content version did not change: %q", changed)
	}
}
