package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRenderEditorialManifest(t *testing.T) {
	project := &Project{
		Name:      "Example",
		Slug:      "example",
		Type:      "Desktop",
		RBVersion: "2026.02",
		Config: map[string]string{
			"MajorVersion": "1",
			"MinorVersion": "2",
			"SubVersion":   "3",
		},
		AllContainers: []*Container{
			{
				Name:  "MainWindow",
				FQN:   "MainWindow",
				Kind:  KindPage,
				Super: "DesktopWindow",
				Members: []Member{
					&Method{baseMember: baseMember{Name: "Opening"}},
				},
			},
		},
	}

	outDir := t.TempDir()
	if err := renderEditorialManifest(project, outDir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "data", "project.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest editorialManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Project.Version != "1.2.3" {
		t.Fatalf("version = %q", manifest.Project.Version)
	}
	if len(manifest.Entities) != 1 || manifest.Entities[0].Location != "pages/mainwindow/" {
		t.Fatalf("entities = %#v", manifest.Entities)
	}
}

func TestRenderEditorialManifestInfersPageSuperclass(t *testing.T) {
	project := &Project{
		Name: "Web Example",
		Slug: "web-example",
		Type: "Web2",
		AllContainers: []*Container{
			{
				Name: "LoginPage",
				FQN:  "LoginPage",
				Kind: KindPage,
				Controls: []*Control{
					{Type: "WebPage", Name: "LoginPage"},
				},
			},
		},
	}

	outDir := t.TempDir()
	if err := renderEditorialManifest(project, outDir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "data", "project.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest editorialManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if got := manifest.Entities[0].SuperName; got != "WebPage" {
		t.Fatalf("inferred page superclass = %q, want WebPage", got)
	}
}
