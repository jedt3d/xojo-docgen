package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var (
		root           = flag.String("root", "", "root dir to scan for *.xojo_project (recursive). Processes each as a separate doc set.")
		single         = flag.String("single", "", "path to a single .xojo_project file to process.")
		out            = flag.String("out", "docs/api", "output dir for generated Markdown (per-project subdirs created beneath).")
		docsRoot       = flag.String("docs", "", "path to the Xojo Documentation dir (for objects.inv). Auto-detected if empty.")
		noLinks        = flag.Bool("no-links", false, "disable external links to the official Xojo docs.")
		includePrivate = flag.Bool("include-private", true, "include private members (collapsed).")
		verbose        = flag.Bool("v", false, "verbose output.")
		publishPrep    = flag.Bool("publish-prep", false, "after generating, write .nojekyll into each docs/api-published/<slug>/ so sites are GitHub-Pages ready.")
		excludeFolder  = flag.String("exclude-folder", "", "comma-separated Xojo Folder item names whose complete manifest subtrees are omitted.")
	)
	flag.Parse()

	if *root == "" && *single == "" {
		// Default to the sample_project dir if present.
		candidate := filepath.Join("tools", "sample_project")
		if _, err := os.Stat(candidate); err == nil {
			*root = candidate
		} else {
			fmt.Fprintln(os.Stderr, "error: provide -root <dir> or -single <project.xojo_project>")
			flag.Usage()
			os.Exit(2)
		}
	}

	// Resolve the Xojo docs root for the link map.
	xojoDocs := *docsRoot
	if xojoDocs == "" && !*noLinks {
		xojoDocs = defaultDocsRoot()
	}
	var lm *LinkMap
	var err error
	if *noLinks {
		lm = &LinkMap{}
	} else {
		lm, err = LoadLinkMap(xojoDocs)
		if err != nil {
			warnf("link map: %v (proceeding without external links)", err)
			lm = &LinkMap{}
		}
	}
	if *verbose && lm.Loaded() {
		fmt.Fprintf(os.Stderr, "link map: %d entries from %s\n", lm.Count(), xojoDocs)
	}

	// Collect projects to process.
	var projects []string
	if *single != "" {
		projects = append(projects, *single)
	} else {
		projects = findProjects(*root)
	}
	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "error: no .xojo_project files found")
		os.Exit(1)
	}

	assets := &templateAssets{baseConfigName: "mkdocs.base.yml"}
	var failed int
	var slugs []string
	for _, projPath := range projects {
		slug, err := processProject(projPath, *out, lm, *includePrivate, *verbose, assets, splitCommaList(*excludeFolder))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", projPath, err)
			failed++
			continue
		}
		slugs = append(slugs, slug)
	}
	// Optionally prep published dirs for static hosting.
	if *publishPrep {
		publishedRoot := filepath.Join(filepath.Dir(*out), "api-published")
		for _, slug := range slugs {
			dir := filepath.Join(publishedRoot, slug)
			if err := renderNoJekyll(dir); err != nil {
				warnf("publish-prep: %s: %v", slug, err)
			}
		}
		if *verbose {
			fmt.Fprintf(os.Stderr, "publish-prep: wrote .nojekyll to %d dir(s) under %s\n", len(slugs), publishedRoot)
		}
	}
	if failed > 0 {
		fmt.Fprintf(os.Stderr, "\n%d project(s) failed.\n", failed)
		os.Exit(1)
	}
}

// findProjects walks root and returns every .xojo_project path.
func findProjects(root string) []string {
	var out []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".xojo_project") {
			out = append(out, path)
		}
		return nil
	})
	return out
}

// processProject parses one project, builds its model, and renders Markdown.
// Returns the project slug.
func processProject(projPath string, outRoot string, lm *LinkMap, includePrivate bool, verbose bool, assets *templateAssets, excludedFolders []string) (string, error) {
	projPath, _ = filepath.Abs(projPath)
	projectDir := filepath.Dir(projPath)
	config, items, err := parseManifest(projPath)
	if err != nil {
		return "", fmt.Errorf("parse manifest: %w", err)
	}

	slug := projectSlug(projPath)
	var excludedCount int
	items, excludedCount = excludeFolderSubtrees(items, excludedFolders)
	if verbose && len(excludedFolders) > 0 {
		fmt.Fprintf(os.Stderr, "project %s: excluded %d manifest item(s) under folder(s): %s\n",
			slug, excludedCount, strings.Join(excludedFolders, ", "))
	}
	p := &Project{
		Name:         projectDisplayName(projPath),
		Slug:         slug,
		Type:         config["Type"],
		RBVersion:    config["RBProjectVersion"],
		Config:       config,
		ManifestPath: projPath,
		ProjectDir:   projectDir,
		ItemsByID:    map[uint64]*Container{},
	}

	// Build container nodes from manifest items.
	for _, it := range items {
		c := &Container{
			ItemType:     it.ItemType,
			Name:         it.Name,
			Kind:         kindFor(it.ItemType),
			ManifestItem: it,
			SourceFile:   it.Path,
		}
		// Apply kind-specific scope defaults.
		p.ItemsByID[it.ID] = c
		p.AllContainers = append(p.AllContainers, c)
	}
	// Link parents and build the tree + FQNs.
	for _, it := range items {
		c := p.ItemsByID[it.ID]
		if c == nil {
			continue
		}
		if it.ParentID != 0 {
			parent := p.ItemsByID[it.ParentID]
			if parent != nil {
				c.Parent = parent
				parent.Children = append(parent.Children, c)
			}
		} else {
			p.RootItems = append(p.RootItems, c)
		}
	}
	// Compute FQNs by walking the parent chain.
	for _, c := range p.AllContainers {
		c.FQN = computeFQN(c)
	}

	// Parse each item's source file (if it has one we recognize).
	for _, c := range p.AllContainers {
		if !shouldDocument(c.Kind) {
			continue
		}
		itemPath := resolveItemPath(projectDir, c.ManifestItem.Path)
		if itemPath == "" || !fileExists(itemPath) {
			continue
		}
		if !isParsableFile(itemPath) {
			continue
		}
		if err := parseFile(itemPath, c); err != nil {
			warnf("%s: %s: %v", slug, c.Name, err)
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "project %s (%s): %d documented items\n", slug, p.Type, countDocumented(p))
	}

	// Render.
	if err := renderMarkdown(p, outRoot, lm, includePrivate, assets); err != nil {
		return slug, err
	}
	return slug, nil
}

func splitCommaList(raw string) []string {
	var values []string
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

// computeFQN builds the fully-qualified name by walking the parent chain.
// Folder parents contribute NO segment; Module/Library/Class contribute theirs.
func computeFQN(c *Container) string {
	var segs []string
	for cur := c; cur != nil; cur = cur.Parent {
		if cur.Kind == KindFolder {
			continue
		}
		segs = append(segs, cur.Name)
	}
	// reverse
	for i, j := 0, len(segs)-1; i < j; i, j = i+1, j-1 {
		segs[i], segs[j] = segs[j], segs[i]
	}
	return strings.Join(segs, ".")
}

func countDocumented(p *Project) int {
	n := 0
	for _, c := range p.AllContainers {
		if shouldDocument(c.Kind) {
			n++
		}
	}
	return n
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// isParsableFile reports whether a path is one we parse for code (vs. binary
// asset files like .xojo_image / .xojo_resources / .xojo_color).
func isParsableFile(p string) bool {
	ext := strings.ToLower(filepath.Ext(p))
	switch ext {
	case ".xojo_code", ".xojo_window", ".xojo_menu", ".xojo_toolbar":
		return true
	}
	return false
}
