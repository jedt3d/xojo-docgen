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
				Name:     "MainWindow",
				FQN:      "MainWindow",
				Kind:     KindPage,
				ItemType: "DesktopWindow",
				Super:    "DesktopWindow",
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
	entity := manifest.Entities[0]
	if entity.LocalName != "MainWindow" || entity.ItemType != "DesktopWindow" {
		t.Fatalf("entity identity = %#v", entity)
	}
	if entity.Navigation.Category != "surface" || entity.Navigation.Section != "Surfaces" {
		t.Fatalf("entity navigation = %#v", entity.Navigation)
	}
}

func TestRenderEditorialManifestInfersPageSuperclass(t *testing.T) {
	project := &Project{
		Name: "Web Example",
		Slug: "web-example",
		Type: "Web2",
		AllContainers: []*Container{
			{
				Name:     "LoginPage",
				FQN:      "LoginPage",
				Kind:     KindPage,
				ItemType: "WebView",
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

func TestEditorialNavigationPreservesLibraryAndModuleHierarchy(t *testing.T) {
	library := &Container{Name: "JinjaXLib", FQN: "JinjaXLib", Kind: KindLibrary}
	module := &Container{
		Name:     "JinjaX",
		FQN:      "JinjaXLib.JinjaX",
		Kind:     KindModule,
		ItemType: "Module",
		Parent:   library,
	}
	class := &Container{
		Name:     "Template",
		FQN:      "JinjaXLib.JinjaX.Template",
		Kind:     KindClass,
		ItemType: "Class",
		Parent:   module,
	}
	project := &Project{
		Name:          "Library Example",
		Slug:          "library-example",
		Type:          "Desktop",
		AllContainers: []*Container{library, module, class},
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
	if len(manifest.Entities) != 2 {
		t.Fatalf("entities = %#v", manifest.Entities)
	}
	for _, entity := range manifest.Entities {
		if entity.Library != "JinjaXLib" || entity.Module != "JinjaXLib.JinjaX" {
			t.Fatalf("hierarchy for %s = library %q, module %q", entity.Name, entity.Library, entity.Module)
		}
		if entity.Navigation.Category != "library" {
			t.Fatalf("category for %s = %q", entity.Name, entity.Navigation.Category)
		}
	}
}

func TestEditorialNavigationSeparatesDialogsAndContainers(t *testing.T) {
	dialog := &Container{
		Name:     "ConfirmDialog",
		FQN:      "ConfirmDialog",
		Kind:     KindPage,
		ItemType: "WebView",
	}
	container := &Container{
		Name:     "CustomerCard",
		FQN:      "CustomerCard",
		Kind:     KindPage,
		ItemType: "WebContainer",
	}

	dialogNavigation := editorialNavigationFor(dialog, "WebDialog", nil, nil)
	if dialogNavigation.Category != "surface" || dialogNavigation.Section != "Dialogs" {
		t.Fatalf("dialog navigation = %#v", dialogNavigation)
	}
	containerNavigation := editorialNavigationFor(container, "WebContainer", nil, nil)
	if containerNavigation.Category != "class" || containerNavigation.Section != "Containers" {
		t.Fatalf("container navigation = %#v", containerNavigation)
	}
}
