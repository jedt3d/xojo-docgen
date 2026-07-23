package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type editorialManifest struct {
	Project   editorialProject    `json:"project"`
	Entities  []editorialEntity   `json:"entities"`
	Databases []editorialDatabase `json:"databases,omitempty"`
}

type editorialProject struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Type      string `json:"type"`
	Xojo      string `json:"xojo"`
	Version   string `json:"version"`
	BundleID  string `json:"bundleId,omitempty"`
	DebugPort string `json:"debugPort,omitempty"`
}

type editorialEntity struct {
	Name       string              `json:"name"`
	LocalName  string              `json:"localName"`
	Kind       string              `json:"kind"`
	ItemType   string              `json:"itemType"`
	Members    int                 `json:"members"`
	SuperName  string              `json:"superName"`
	Location   string              `json:"location"`
	Library    string              `json:"library,omitempty"`
	Module     string              `json:"module,omitempty"`
	Navigation editorialNavigation `json:"navigation"`
}

type editorialNavigation struct {
	Category string `json:"category"`
	Section  string `json:"section"`
}

type editorialDatabase struct {
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Dialect       string `json:"dialect"`
	Source        string `json:"source"`
	Tables        int    `json:"tables"`
	Columns       int    `json:"columns"`
	Relationships int    `json:"relationships"`
	Views         int    `json:"views"`
	Triggers      int    `json:"triggers"`
	Location      string `json:"location"`
}

func renderEditorialManifest(project *Project, outDir string) error {
	manifest := editorialManifest{
		Project: editorialProject{
			Name:      project.Name,
			Slug:      project.Slug,
			Type:      project.Type,
			Xojo:      project.RBVersion,
			Version:   projectVersion(project.Config),
			BundleID:  project.Config["OSXBundleID"],
			DebugPort: project.Config["WebDebugPort"],
		},
	}

	for _, container := range project.AllContainers {
		if !shouldDocument(container.Kind) {
			continue
		}
		superName := container.Super
		if superName == "" && container.Kind == KindPage && len(container.Controls) > 0 {
			superName = container.Controls[0].Type
		}
		if superName == "" {
			superName = "—"
		}
		library := nearestContainer(container, KindLibrary)
		module := nearestContainer(container, KindModule)
		manifest.Entities = append(manifest.Entities, editorialEntity{
			Name:       container.FQN,
			LocalName:  container.Name,
			Kind:       editorialKind(container.Kind),
			ItemType:   container.ItemType,
			Members:    len(container.Members),
			SuperName:  superName,
			Location:   strings.TrimSuffix(containerFilePath(container), ".md") + "/",
			Library:    containerName(library),
			Module:     containerName(module),
			Navigation: editorialNavigationFor(container, superName, library, module),
		})
	}
	sort.Slice(manifest.Entities, func(first int, second int) bool {
		return manifest.Entities[first].Name < manifest.Entities[second].Name
	})
	for _, database := range project.Databases {
		manifest.Databases = append(manifest.Databases, editorialDatabase{
			Name:          database.Name,
			Slug:          database.Slug,
			Dialect:       database.Dialect,
			Source:        database.Source,
			Tables:        len(database.Tables),
			Columns:       countDatabaseColumns(database),
			Relationships: countDatabaseRelationships(database),
			Views:         len(database.Views),
			Triggers:      len(database.Triggers),
			Location:      databaseDictionaryLocation(database.Slug, ""),
		})
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dataDir := filepath.Join(outDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "project.json"), data, 0o644)
}

func projectVersion(config map[string]string) string {
	parts := []string{config["MajorVersion"], config["MinorVersion"], config["SubVersion"]}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, ".")
}

func editorialKind(kind ContainerKind) string {
	switch kind {
	case KindPage:
		return "Page"
	case KindWebSession:
		return "Session"
	default:
		return kindLabel(kind)
	}
}

func nearestContainer(container *Container, kind ContainerKind) *Container {
	for current := container; current != nil; current = current.Parent {
		if current.Kind == kind {
			return current
		}
	}
	return nil
}

func containerName(container *Container) string {
	if container == nil {
		return ""
	}
	return container.FQN
}

func editorialNavigationFor(
	container *Container,
	superName string,
	library *Container,
	module *Container,
) editorialNavigation {
	isContainer := strings.Contains(strings.ToLower(container.ItemType), "container") ||
		strings.Contains(strings.ToLower(superName), "container")
	if library != nil || module != nil {
		return editorialNavigation{Category: "library", Section: "Library"}
	}

	switch container.Kind {
	case KindPage:
		if isContainer {
			return editorialNavigation{Category: "class", Section: "Containers"}
		}
		if strings.Contains(strings.ToLower(superName), "dialog") ||
			strings.HasSuffix(strings.ToLower(container.Name), "dialog") {
			return editorialNavigation{Category: "surface", Section: "Dialogs"}
		}
		return editorialNavigation{Category: "surface", Section: "Surfaces"}
	case KindClass:
		if isContainer {
			return editorialNavigation{Category: "class", Section: "Containers"}
		}
		return editorialNavigation{Category: "class", Section: "Classes"}
	case KindInterface:
		return editorialNavigation{Category: "class", Section: "Interfaces"}
	case KindWebSession:
		return editorialNavigation{Category: "class", Section: "Sessions"}
	case KindModule:
		return editorialNavigation{Category: "library", Section: "Modules"}
	case KindMenuBar:
		return editorialNavigation{Category: "misc", Section: "Menu Bars"}
	case KindToolbar:
		return editorialNavigation{Category: "misc", Section: "Toolbars"}
	default:
		return editorialNavigation{Category: "misc", Section: "Other"}
	}
}
