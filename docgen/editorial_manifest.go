package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type editorialManifest struct {
	Project  editorialProject  `json:"project"`
	Entities []editorialEntity `json:"entities"`
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
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Members   int    `json:"members"`
	SuperName string `json:"superName"`
	Location  string `json:"location"`
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
		manifest.Entities = append(manifest.Entities, editorialEntity{
			Name:      container.FQN,
			Kind:      editorialKind(container.Kind),
			Members:   len(container.Members),
			SuperName: superName,
			Location:  strings.TrimSuffix(containerFilePath(container), ".md") + "/",
		})
	}
	sort.Slice(manifest.Entities, func(first int, second int) bool {
		return manifest.Entities[first].Name < manifest.Entities[second].Name
	})

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
