package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// renderMarkdown walks the project and writes one .md per documented container
// under outDir/<slug>/, plus an index.md landing page and a featured image.
func renderMarkdown(p *Project, outRoot string, lm *LinkMap, includePrivate bool, assets *templateAssets) error {
	outDir := filepath.Join(outRoot, p.Slug)
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clear generated project docs: %w", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	// Featured images: a landscape placeholder (1200x630) for desktop/wide
	// screens and a portrait placeholder (800x1000) for phones and portrait
	// tablets. The landing page swaps between them with a <picture> element.
	// Users replace these files with real artwork later.
	assetsDir := filepath.Join(outDir, "assets")
	featuredPath := filepath.Join(assetsDir, "featured.png")
	if err := generateFeaturedPNG(featuredPath); err != nil {
		warnf("%s: featured image: %v", p.Slug, err)
	}
	featuredPortraitPath := filepath.Join(assetsDir, "featured_portrait.png")
	if err := generateFeaturedPortraitPNG(featuredPortraitPath); err != nil {
		warnf("%s: featured portrait image: %v", p.Slug, err)
	}

	// Shared green theme stylesheet (so each per-project site is self-contained).
	cssPath := filepath.Join(outDir, "stylesheets", "extra.css")
	if err := os.MkdirAll(filepath.Dir(cssPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cssPath, []byte(xojoGreenCSS), 0o644); err != nil {
		warnf("%s: theme css: %v", p.Slug, err)
	}

	// Client-side JS: Prism core + Xojo grammar + fullscreen modal. Vendored
	// via go:embed so each site is fully self-contained (no CDN dependency).
	jsDir := filepath.Join(outDir, "javascripts")
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		return err
	}
	for name, content := range map[string]string{
		"prism.js":           prismCoreJS,
		"xojo.prism.js":      xojoPrismJS,
		"source-modal.js":    sourceModalJS,
		"landing-sidebar.js": landingSidebarJS,
	} {
		if err := os.WriteFile(filepath.Join(jsDir, name), []byte(content), 0o644); err != nil {
			warnf("%s: js %s: %v", p.Slug, name, err)
		}
	}

	// Landing page.
	if err := renderLandingPage(p, outDir, lm); err != nil {
		return err
	}

	// Per-container pages.
	rc := &renderCtx{
		proj:           p,
		lm:             lm,
		includePrivate: includePrivate,
		assets:         assets,
		internalTypes:  buildInternalTypeMap(p),
	}
	for _, c := range p.AllContainers {
		if !shouldDocument(c.Kind) {
			continue
		}
		if err := renderContainerPage(c, outDir, rc); err != nil {
			warnf("%s: %s: %v", p.Slug, c.FQN, err)
		}
	}

	// Per-project mkdocs.yml.
	if err := renderProjectMkdocsYml(p, outDir); err != nil {
		return err
	}
	return nil
}

// templateAssets carries shared config the renderer needs.
type templateAssets struct {
	baseConfigName string // name of the base mkdocs config (mkdocs.base.yml)
}

// renderCtx bundles per-render options.
type renderCtx struct {
	proj           *Project
	lm             *LinkMap
	includePrivate bool
	assets         *templateAssets
	internalTypes  map[string]string // lowercase simple-name -> path to its page, relative to project doc root
	currentPath    string            // path of the page currently being rendered (relative to project doc root); used to compute cross-folder links
}

// buildInternalTypeMap indexes every documented container in the project by
// its simple name (lowercased) so user-defined types in signatures can link to
// their sibling pages. e.g. Customer -> "classes/customer.md".
func buildInternalTypeMap(p *Project) map[string]string {
	m := map[string]string{}
	for _, c := range p.AllContainers {
		if !shouldDocument(c.Kind) {
			continue
		}
		key := strings.ToLower(c.Name)
		if key == "" {
			continue
		}
		rel := containerFilePath(c)
		// First declaration wins; later duplicates (same simple name in
		// different namespaces) keep the first to keep links stable.
		if _, exists := m[key]; !exists {
			m[key] = rel
		}
	}
	return m
}

// ---- landing page (index.md) ----

func renderLandingPage(p *Project, outDir string, lm *LinkMap) error {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", p.Name))

	// Project facts + Contents counts are rendered into a card in the LEFT
	// sidebar by landing-sidebar.js (not in the center column). We emit them
	// here as a hidden JSON payload so the script has the data to clone.
	if metaJSON, err := landingMetaJSON(p); err == nil {
		fmt.Fprintf(&b, "<script type=\"application/json\" id=\"lp-meta\">%s</script>\n\n", metaJSON)
	}

	// Featured image: a responsive <picture> sits directly above the Entities
	// block. Landscape (featured.png) for wide screens, portrait
	// (featured_portrait.png) for phones/portrait tablets via art direction.
	// Users replace the PNGs in assets/ with real artwork.
	b.WriteString("<!-- Featured image: replace assets/featured.png (landscape, 1200x630) and assets/featured_portrait.png (portrait, 800x1000). -->\n")
	fmt.Fprintf(&b, "<picture class=\"featured\">\n")
	fmt.Fprintf(&b, "  <source media=\"(max-width: 768px)\" srcset=\"assets/featured_portrait.png\">\n")
	fmt.Fprintf(&b, "  <img src=\"assets/featured.png\" alt=\"%s — featured\">\n", htmlEscape(p.Name))
	b.WriteString("</picture>\n\n")

	// Entity table.
	b.WriteString("## Entities\n\n")
	b.WriteString("| Name | Kind | Members | Super |\n|---|---|---:|---|\n")
	rows := landingRows(p)
	for _, r := range rows {
		fmt.Fprintf(&b, "| [%s](%s) | %s | %d | %s |\n",
			r.fqn, r.link, r.kind, r.members, r.super)
	}
	b.WriteString("\n")

	// Official docs pointer.
	if lm.Loaded() {
		fmt.Fprintf(&b, "> Entity types that exist in the official Xojo API link to [%s](%s).\n\n",
			"the Xojo documentation", "https://documentation.xojo.com/index.html")
	}

	return os.WriteFile(filepath.Join(outDir, "index.md"), []byte(b.String()), 0o644)
}

// landingMeta is the payload emitted as <script id="lp-meta"> on the landing
// page and consumed by landing-sidebar.js to build the sidebar card.
type landingMeta struct {
	Title    string            `json:"title"`
	Facts    []landingMetaFact `json:"facts"`
	Contents []landingMetaKind `json:"contents"`
}

type landingMetaFact struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type landingMetaKind struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

// landingMetaJSON builds the sidebar payload for the landing page and returns
// it JSON-encoded (safe to embed in a <script type="application/json"> block).
func landingMetaJSON(p *Project) (string, error) {
	meta := landingMeta{Title: p.Name, Facts: projectFacts(p)}
	counts := countByKind(p)
	for _, k := range sortedKindKeys(counts) {
		meta.Contents = append(meta.Contents, landingMetaKind{Kind: k, Count: counts[k]})
	}
	out, err := jsonMarshalHTMLSafe(meta)
	if err != nil {
		return "", err
	}
	return out, nil
}

// projectFacts collects the project's key/value metadata in display order.
func projectFacts(p *Project) []landingMetaFact {
	var facts []landingMetaFact
	add := func(label, v string) {
		if v != "" {
			facts = append(facts, landingMetaFact{Label: label, Value: v})
		}
	}
	add("Type", p.Type)
	add("RBProjectVersion", p.RBVersion)
	if v, ok := p.Config["OSXBundleID"]; ok {
		add("Bundle ID", v)
	}
	if v, ok := p.Config["MajorVersion"]; ok {
		ver := v
		if mn, ok := p.Config["MinorVersion"]; ok {
			ver += "." + mn
		}
		if sv, ok := p.Config["SubVersion"]; ok {
			ver += "." + sv
		}
		add("Version", ver)
	}
	if v, ok := p.Config["DefaultWindow"]; ok {
		add("Default window", v)
	}
	if v, ok := p.Config["DefaultScreen"]; ok {
		add("Default screen", v)
	}
	if v, ok := p.Config["DefaultMobileView"]; ok {
		add("Default mobile view", v)
	}
	if v, ok := p.Config["AppMenuBar"]; ok {
		add("App menu bar", v)
	}
	if v, ok := p.Config["WebDebugPort"]; ok {
		add("Web debug port", v)
	}
	if v, ok := p.Config["IsWebProject"]; ok {
		add("Is web project", v)
	}
	return facts
}

// jsonMarshalHTMLSafe marshals v to compact JSON and escapes <, >, & so the
// result can be embedded inside <script type="application/json"> without
// terminating the tag early on content like "<".
func jsonMarshalHTMLSafe(v any) (string, error) {
	out, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	r := strings.NewReplacer(
		"<", "\\u003c",
		">", "\\u003e",
		"&", "\\u0026",
	)
	return r.Replace(string(out)), nil
}

type landingRow struct {
	fqn     string
	link    string
	kind    string
	members int
	super   string
}

func landingRows(p *Project) []landingRow {
	var rows []landingRow
	for _, c := range p.AllContainers {
		if !shouldDocument(c.Kind) {
			continue
		}
		link := containerLink(c, p.Slug)
		super := c.Super
		if super == "" {
			super = "—"
		}
		rows = append(rows, landingRow{
			fqn:     c.FQN,
			link:    link,
			kind:    kindLabel(c.Kind),
			members: len(c.Members),
			super:   super,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].fqn < rows[j].fqn })
	return rows
}

// containerLink returns the relative link to a container's page (from the
// project landing). Pages are organized into subfolders by kind.
func containerLink(c *Container, slug string) string {
	return "./" + containerFilePath(c)
}

// containerFilePath returns the path (relative to the project dir) where a
// container's Markdown page is written.
func containerFilePath(c *Container) string {
	sub := kindSubdir(c.Kind)
	name := slugify(c.FQN) + ".md"
	if sub == "" {
		return name
	}
	return sub + "/" + name
}

func kindSubdir(k ContainerKind) string {
	switch k {
	case KindClass, KindModule, KindInterface, KindWebSession:
		return "classes"
	case KindPage:
		return "pages"
	case KindMenuBar:
		return "menus"
	case KindToolbar:
		return "toolbars"
	}
	return "misc"
}

func countByKind(p *Project) map[string]int {
	m := map[string]int{}
	for _, c := range p.AllContainers {
		if !shouldDocument(c.Kind) {
			continue
		}
		m[kindLabel(c.Kind)]++
	}
	return m
}

func sortedKindKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// kindLabel returns a human label for a ContainerKind.
func kindLabel(k ContainerKind) string {
	switch k {
	case KindClass:
		return "Class"
	case KindModule:
		return "Module"
	case KindInterface:
		return "Interface"
	case KindWebSession:
		return "Session"
	case KindPage:
		return "Page / Window"
	case KindMenuBar:
		return "Menu Bar"
	case KindToolbar:
		return "Toolbar"
	case KindFolder:
		return "Folder"
	case KindLibrary:
		return "Library"
	case KindBuildSteps:
		return "Build Steps"
	}
	return "Other"
}

// ---- per-container page ----

func renderContainerPage(c *Container, outDir string, rc *renderCtx) error {
	rel := containerFilePath(c)
	full := filepath.Join(outDir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	// Record the page's own path so cross-folder internal links compute the
	// right number of "../" hops. Saved/restored because rc is shared.
	prevPath := rc.currentPath
	rc.currentPath = rel
	defer func() { rc.currentPath = prevPath }()

	var b strings.Builder

	// Frontmatter (MkDocs-Material accepts this).
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", c.FQN)
	b.WriteString("---\n\n")

	// Title + badges.
	fmt.Fprintf(&b, "# %s\n\n", c.FQN)
	var badges []string
	badges = append(badges, fmt.Sprintf("`%s`", kindLabel(c.Kind)))
	badges = append(badges, fmt.Sprintf("`%s`", c.Scope.String()))
	if dep, msg := containerDeprecated(c); dep {
		m := "Deprecated"
		if msg != "" {
			m += ": " + msg
		}
		badges = append(badges, "`⚠ "+m+"`")
	}
	b.WriteString(strings.Join(badges, " "))
	b.WriteString("\n\n")

	// Inheritance.
	if c.Super != "" {
		b.WriteString("**Inherits:** " + renderTypeRef(c.Super, rc) + "  \n")
	}
	if len(c.Implements) > 0 {
		var refs []string
		for _, i := range c.Implements {
			refs = append(refs, renderTypeRef(i, rc))
		}
		b.WriteString("**Implements:** " + strings.Join(refs, ", ") + "  \n")
	}
	if c.Super != "" || len(c.Implements) > 0 {
		b.WriteString("\n")
	}

	// Description.
	if d := containerDocs(c); d != "" {
		b.WriteString(d + "\n\n")
	}

	// Named notes as labeled subsections.
	for _, n := range c.NamedNotes {
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", n.Name, strings.TrimSpace(n.Body))
	}

	// Controls (for pages/windows).
	if len(c.Controls) > 0 {
		b.WriteString("## Controls\n\n")
		for _, ctrl := range c.Controls {
			renderControlTree(&b, ctrl, rc, 0)
		}
		b.WriteString("\n")
	}

	// Members, grouped by kind.
	membersByKind := groupMembers(c.Members, rc.includePrivate)
	order := []string{
		"Event Definition", "Event Handler",
		"Method", "Delegate",
		"Property", "Computed Property",
		"Constant", "Enum",
	}
	rendered := map[string]bool{}
	for _, kind := range order {
		if ms, ok := membersByKind[kind]; ok && len(ms) > 0 {
			renderMemberGroup(&b, kind, ms, rc)
			rendered[kind] = true
		}
	}
	// any remaining kinds
	for _, m := range c.Members {
		k := m.MemberKind()
		if rendered[k] {
			continue
		}
		rendered[k] = true
		var ms []Member
		for _, mm := range c.Members {
			if mm.MemberKind() == k {
				if !rc.includePrivate && mm.MemberScope() == ScopePrivate {
					continue
				}
				ms = append(ms, mm)
			}
		}
		if len(ms) > 0 {
			renderMemberGroup(&b, k, ms, rc)
		}
	}

	return os.WriteFile(full, []byte(b.String()), 0o644)
}

// groupMembers partitions members by kind, optionally hiding Private.
func groupMembers(members []Member, includePrivate bool) map[string][]Member {
	out := map[string][]Member{}
	for _, m := range members {
		if !includePrivate && m.MemberScope() == ScopePrivate {
			continue
		}
		k := m.MemberKind()
		out[k] = append(out[k], m)
	}
	return out
}

// renderMemberGroup writes one section for a member kind, with Private members
// collapsed under an <details> block.
func renderMemberGroup(b *strings.Builder, kind string, ms []Member, rc *renderCtx) {
	// Split public+protected vs private (within the group).
	var pubProt, priv []Member
	for _, m := range ms {
		if m.MemberScope() == ScopePrivate {
			priv = append(priv, m)
		} else {
			pubProt = append(pubProt, m)
		}
	}
	if len(pubProt) > 0 {
		fmt.Fprintf(b, "## %s\n\n", pluralizeKind(kind))
		for _, m := range pubProt {
			renderMember(b, m, rc)
		}
	}
	if len(priv) > 0 {
		fmt.Fprintf(b, "### %s — internal\n\n", pluralizeKind(kind))
		b.WriteString("<details class=\"internal\"><summary>Private / internal members</summary>\n\n")
		for _, m := range priv {
			renderMember(b, m, rc)
		}
		b.WriteString("</details>\n\n")
	}
}

func pluralizeKind(k string) string {
	switch k {
	case "Computed Property":
		return "Computed Properties"
	case "Property":
		return "Properties"
	case "Constant":
		return "Constants"
	case "Enum":
		return "Enums"
	case "Delegate":
		return "Delegates"
	case "Event Definition":
		return "Event Definitions"
	case "Event Handler":
		return "Event Handlers"
	}
	return k + "s"
}

func renderMember(b *strings.Builder, m Member, rc *renderCtx) {
	switch v := m.(type) {
	case *Method:
		renderMethod(b, v, rc)
	case *Property:
		renderProperty(b, v, rc)
	case *ComputedProperty:
		renderComputedProperty(b, v, rc)
	case *Constant:
		renderConstant(b, v, rc)
	case *Enum:
		renderEnum(b, v, rc)
	case *Delegate:
		renderDelegate(b, v, rc)
	case *EventDef:
		renderEventDef(b, v, rc)
	case *EventHandler:
		renderEventHandler(b, v, rc)
	}
}

func renderMethod(b *strings.Builder, m *Method, rc *renderCtx) {
	fmt.Fprintf(b, "### %s\n\n", m.Name)
	writeScopeBadge(b, m.Scope, m.IsShared)
	if m.IsDeprecated {
		writeDeprecated(b, m.DeprecMsg)
	}
	// Signature with linked types, rendered as HTML so <a> links work.
	sig := linkifySignatureHTML(m.Signature, rc)
	writeCodeBlock(b, sig)
	if d := memberDocs(m); d != "" {
		b.WriteString(d + "\n\n")
	}
	writeSourceDetails(b, m.Source)
}

func renderProperty(b *strings.Builder, p *Property, rc *renderCtx) {
	fmt.Fprintf(b, "### %s\n\n", p.Name)
	writeScopeBadge(b, p.Scope, p.IsShared)
	// Build a clean declaration WITHOUT the scope keyword (Private/Public/Protected),
	// which is already shown in the badge above. Reconstruct from parsed name+type
	// so the VB/Xojo body keyword doesn't leak into the rendered code.
	decl := fmt.Sprintf("%s As %s", p.Name, p.Type)
	if p.DefaultValue != "" {
		decl += " = " + p.DefaultValue
	}
	writeCodeBlock(b, linkifySignatureHTML(decl, rc))
	if d := memberDocs(p); d != "" {
		b.WriteString(d + "\n\n")
	}
}

func renderComputedProperty(b *strings.Builder, c *ComputedProperty, rc *renderCtx) {
	fmt.Fprintf(b, "### %s\n\n", c.Name)
	writeScopeBadge(b, c.Scope, c.IsShared)
	ro := ""
	if c.IsReadOnly {
		ro = " (read-only)"
	}
	writeCodeBlock(b, fmt.Sprintf("%s As %s%s", htmlEscape(c.Name), linkifyTypeHTML(c.Type, rc), ro))
	if d := memberDocs(c); d != "" {
		b.WriteString(d + "\n\n")
	}
	// Source = the Get/Set accessor bodies combined.
	src := strings.TrimSpace(joinNonEmpty("\n\n", c.GetterSrc, c.SetterSrc))
	writeSourceDetails(b, src)
}

// joinNonEmpty joins non-empty strings with sep, skipping empty parts.
func joinNonEmpty(sep string, parts ...string) string {
	var out []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}

func renderConstant(b *strings.Builder, c *Constant, rc *renderCtx) {
	fmt.Fprintf(b, "### %s\n\n", c.Name)
	writeScopeBadge(b, c.Scope, false)
	dyn := ""
	if c.Dynamic {
		dyn = ", Dynamic"
	}
	def := c.Default
	def = strings.ReplaceAll(def, "\r", " ")
	def = strings.ReplaceAll(def, "\n", " ")
	writeCodeBlock(b, fmt.Sprintf("Const %s As %s%s = %s",
		htmlEscape(c.Name), htmlEscape(c.Type), dyn, htmlEscape(quoteIfNeeded(def))))
}

func renderEnum(b *strings.Builder, e *Enum, rc *renderCtx) {
	fmt.Fprintf(b, "### %s\n\n", e.Name)
	writeScopeBadge(b, e.Scope, e.IsShared)
	var inner strings.Builder
	fmt.Fprintf(&inner, "Enum %s As %s\n", htmlEscape(e.Name), htmlEscape(e.Type))
	for _, m := range e.Members {
		if m.Value != "" {
			fmt.Fprintf(&inner, "  %s = %s\n", htmlEscape(m.Name), htmlEscape(m.Value))
		} else {
			fmt.Fprintf(&inner, "  %s\n", htmlEscape(m.Name))
		}
	}
	inner.WriteString("End Enum")
	writeCodeBlock(b, inner.String())
}

func renderDelegate(b *strings.Builder, d *Delegate, rc *renderCtx) {
	fmt.Fprintf(b, "### %s\n\n", d.Name)
	writeScopeBadge(b, d.Scope, d.IsShared)
	writeCodeBlock(b, linkifySignatureHTML(d.RawDecl, rc))
	writeSourceDetails(b, d.Source)
}

func renderEventDef(b *strings.Builder, e *EventDef, rc *renderCtx) {
	fmt.Fprintf(b, "### %s\n\n", e.Name)
	b.WriteString("`Event Definition`\n\n")
	writeCodeBlock(b, linkifySignatureHTML(e.RawDecl, rc))
	if d := memberDocs(e); d != "" {
		b.WriteString(d + "\n\n")
	}
	writeSourceDetails(b, e.Source)
}

func renderEventHandler(b *strings.Builder, e *EventHandler, rc *renderCtx) {
	title := e.Name
	if e.ControlName != "" {
		title = e.ControlName + "." + e.Name
	}
	fmt.Fprintf(b, "### %s\n\n", title)
	scopeKw := "Public"
	if e.Scope == ScopeProtected {
		scopeKw = "Protected"
	}
	if e.Scope == ScopePrivate {
		scopeKw = "Private"
	}
	fmt.Fprintf(b, "`%s` `Event Handler`", scopeKw)
	b.WriteString("\n\n")
	writeCodeBlock(b, linkifySignatureHTML(e.RawDecl, rc))
	if d := memberDocs(e); d != "" {
		b.WriteString(d + "\n\n")
	}
	writeSourceDetails(b, e.Source)
}

func writeScopeBadge(b *strings.Builder, s Scope, shared bool) {
	kw := s.String()
	if shared {
		kw += ", Shared"
	}
	fmt.Fprintf(b, "`%s`\n\n", kw)
}

func writeDeprecated(b *strings.Builder, msg string) {
	if msg == "" {
		b.WriteString("`⚠ Deprecated`\n\n")
		return
	}
	fmt.Fprintf(b, "`⚠ Deprecated: %s`\n\n", msg)
}

// renderControlTree writes a nested bullet list of the Begin/End UI tree.
func renderControlTree(b *strings.Builder, ctrl *Control, rc *renderCtx, depth int) {
	indent := strings.Repeat("  ", depth)
	typeRef := renderTypeRef(ctrl.Type, rc)
	label := ctrl.Name
	if cap, ok := ctrl.Properties["Caption"]; ok && cap != "" {
		label += fmt.Sprintf(" — %q", cap)
	} else if t, ok := ctrl.Properties["Title"]; ok && t != "" {
		label += fmt.Sprintf(" — %q", t)
	} else if txt, ok := ctrl.Properties["Text"]; ok && txt != "" {
		label += fmt.Sprintf(" — %q", txt)
	}
	fmt.Fprintf(b, "%s- %s **%s**\n", indent, typeRef, label)
	for _, child := range ctrl.Children {
		renderControlTree(b, child, rc, depth+1)
	}
}

// ---- type linking ----

// renderTypeRef returns a MARKDOWN link if name resolves, else the bare name.
// Resolution order: official Xojo docs (objects.inv) first, then project-internal
// types (sibling pages). Use in PROSE contexts; use renderTypeRefHTML in code blocks.
// Internal links use the .md form because MkDocs rewrites Markdown links to
// directory URLs during build.
func renderTypeRef(name string, rc *renderCtx) string {
	if url, ok := rc.lm.Link(name); ok {
		return fmt.Sprintf("[%s](%s)", name, url)
	}
	if rel, ok := rc.internalLinkMD(name); ok {
		return fmt.Sprintf("[%s](%s)", name, rel)
	}
	return name
}

// renderTypeRefHTML returns an HTML <a> link if name resolves, else the
// HTML-escaped bare name. Same resolution order as renderTypeRef. Use INSIDE
// <pre><code> blocks. Internal links use the directory-URL form (ending "/")
// because raw HTML <a href> is NOT rewritten by MkDocs.
func renderTypeRefHTML(name string, rc *renderCtx) string {
	if url, ok := rc.lm.Link(name); ok {
		return fmt.Sprintf("<a href=\"%s\">%s</a>", url, htmlEscape(name))
	}
	if rel, ok := rc.internalLinkDir(name); ok {
		return fmt.Sprintf("<a href=\"%s\">%s</a>", rel, htmlEscape(name))
	}
	return htmlEscape(name)
}

// internalLinkMD returns the .md relative path for a project type, for use in
// Markdown link syntax (which MkDocs rewrites to directory URLs at build time).
func (rc *renderCtx) internalLinkMD(name string) (string, bool) {
	target, ok := rc.lookupInternal(name)
	if !ok {
		return "", false
	}
	return relLinkMD(rc.currentPath, target), true
}

// internalLinkDir returns the directory-URL relative path for a project type
// (ending "/"), for use in raw HTML <a href> (which MkDocs does NOT rewrite).
func (rc *renderCtx) internalLinkDir(name string) (string, bool) {
	target, ok := rc.lookupInternal(name)
	if !ok {
		return "", false
	}
	return relLink(rc.currentPath, target), true
}

// lookupInternal finds the page path for a project-defined type name.
func (rc *renderCtx) lookupInternal(name string) (string, bool) {
	if rc.internalTypes == nil {
		return "", false
	}
	target, ok := rc.internalTypes[strings.ToLower(strings.TrimSpace(name))]
	return target, ok
}

// relLinkMD computes a .md relative path for Markdown links (MkDocs rewrites
// these to directory URLs, so .md form is correct here).
func relLinkMD(fromPage, toPage string) string {
	fromDir := filepath.Dir(fromPage)
	rel, err := filepath.Rel(fromDir, toPage)
	if err != nil {
		return toPage
	}
	rel = filepath.ToSlash(rel)
	return rel
}

// relLink computes a relative hyperlink from a source page to a target page,
// accounting for MkDocs use_directory_urls:true, which serves each page X.md
// as a directory X/ containing index.html. So both the source and target are
// "directories" in the served URL space.
//
// Examples (source -> target => output):
//
//	classes/invoice.md -> classes/customer.md   => "../customer/"   (sibling)
//	classes/invoice.md -> classes/invoice.md    => "./"             (self)
//	pages/screen.md    -> classes/customer.md   => "../classes/customer/"
//
// The output ends with "/" (the directory URL) and never includes ".md",
// because raw HTML <a href> links inside <pre><code> blocks are NOT rewritten
// by MkDocs — we must emit the final served path ourselves.
func relLink(fromPage, toPage string) string {
	// Treat each page as a directory: drop the filename, keep the dir.
	fromDir := filepath.Dir(fromPage)                          // e.g. "classes"
	toBase := strings.TrimSuffix(filepath.Base(toPage), ".md") // e.g. "customer"
	toDir := filepath.Dir(toPage)                              // e.g. "classes"
	// Relative hop from fromDir to toDir, then into the target page's directory.
	relDir, err := filepath.Rel(fromDir, toDir)
	if err != nil {
		return toBase + "/"
	}
	relDir = filepath.ToSlash(relDir)
	// Rel() returns "." for the same directory. For a sibling page, the served
	// URL is one level up from the source page's directory, then into the
	// target directory: "../customer/".
	switch {
	case relDir == ".":
		// Same folder: source is invoice/, target is customer/ -> "../customer/"
		return "../" + toBase + "/"
	case relDir == "..":
		return "../../" + toBase + "/"
	default:
		// Different subfolder: relDir is like "../classes" or "classes".
		// Source page dir (e.g. pages/screen/) needs to hop up to reach it.
		if strings.HasPrefix(relDir, "../") {
			// already goes up; add one more ".." for the source page dir, then target
			return relDir + "/../" + toBase + "/"
		}
		// relDir is a child (rare for cross-folder); hop up from source dir first
		return "../" + relDir + "/" + toBase + "/"
	}
}

// linkifyType links a type token (e.g. "WebButton", "Integer", "String()").
// Returns MARKDOWN-link syntax — use only in prose.
func linkifyType(typ string, rc *renderCtx) string {
	typ = strings.TrimSpace(typ)
	core := typ
	suffix := ""
	if strings.HasSuffix(core, "()") {
		core = strings.TrimSuffix(core, "()")
		suffix = "()"
	}
	return renderTypeRef(core, rc) + suffix
}

// linkifyTypeHTML is the code-block variant of linkifyType — emits <a> tags.
func linkifyTypeHTML(typ string, rc *renderCtx) string {
	typ = strings.TrimSpace(typ)
	core := typ
	suffix := ""
	if strings.HasSuffix(core, "()") {
		core = strings.TrimSuffix(core, "()")
		suffix = "()"
	}
	return renderTypeRefHTML(core, rc) + suffix
}

// linkifySignature takes a signature/code line and turns known type tokens
// into MARKDOWN links. Use only in prose. (See linkifySignatureHTML for code blocks.)
func linkifySignature(sig string, rc *renderCtx) string {
	if sig == "" {
		return ""
	}
	words := strings.Fields(sig)
	for i := 0; i < len(words); i++ {
		if strings.EqualFold(words[i], "As") && i+1 < len(words) {
			i++
			words[i] = linkTypeToken(words[i], rc)
		}
	}
	return strings.Join(words, " ")
}

// linkifySignatureHTML is the code-block variant: emits real <a> tags so links
// work inside <pre><code>. Tokens other than types are HTML-escaped.
func linkifySignatureHTML(sig string, rc *renderCtx) string {
	if sig == "" {
		return ""
	}
	words := strings.Fields(sig)
	for i := 0; i < len(words); i++ {
		// Find "As" tokens and link the NEXT word (the type) as HTML.
		if strings.EqualFold(words[i], "As") && i+1 < len(words) {
			i++
			words[i] = linkTypeTokenHTML(words[i], rc)
		} else {
			// Escape everything else so the code block is safe.
			words[i] = htmlEscape(words[i])
		}
	}
	return strings.Join(words, " ")
}

// linkTypeToken links a single type token (MARKDOWN syntax), preserving
// trailing punctuation like commas, parens, or array "()" suffixes. Uses the
// shared renderTypeRef so official-doc and project-internal links both resolve.
func linkTypeToken(tok string, rc *renderCtx) string {
	core, arrSuffix, suffix := splitTypeToken(tok)
	if !looksLikeType(core) {
		return tok
	}
	return renderTypeRef(core, rc) + arrSuffix + suffix
}

// linkTypeTokenHTML is the code-block variant: <a> tags via renderTypeRefHTML.
func linkTypeTokenHTML(tok string, rc *renderCtx) string {
	core, arrSuffix, suffix := splitTypeToken(tok)
	if !looksLikeType(core) {
		return htmlEscape(tok)
	}
	return renderTypeRefHTML(core, rc) + arrSuffix + suffix
}

// splitTypeToken separates a type token into its core type name, optional
// array "()" suffix, and trailing punctuation (commas, closing parens).
func splitTypeToken(tok string) (core, arrSuffix, suffix string) {
	core = tok
	for len(core) > 0 && (core[len(core)-1] == ',' || core[len(core)-1] == ')') {
		suffix = string(core[len(core)-1]) + suffix
		core = core[:len(core)-1]
	}
	if strings.HasSuffix(core, "()") {
		arrSuffix = "()"
		core = strings.TrimSuffix(core, "()")
	}
	return core, arrSuffix, suffix
}

// htmlEscape escapes the basics for safe rendering inside <pre><code>.
func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return r.Replace(s)
}

// writeCodeBlock writes a <pre><code> HTML block with the given (already
// HTML-linkified) content. Material renders this as a styled code block with
// working links, unlike fenced code blocks which would show [text](url) raw.
func writeCodeBlock(b *strings.Builder, content string) {
	fmt.Fprintf(b, "<pre><code>%s</code></pre>\n\n", content)
}

// writeSourceDetails emits a collapsible "Source" block with the full VB/Xojo
// source. The <pre><code> carries class="language-xojo" so Prism.js highlights
// it client-side. Source is HTML-escaped here (Prism reads textContent, so
// escaping is safe and necessary). Omitted entirely if source is empty.
func writeSourceDetails(b *strings.Builder, source string) {
	source = strings.TrimSpace(source)
	if source == "" {
		return
	}
	b.WriteString("<details class=\"source\"><summary>Source</summary>\n\n")
	fmt.Fprintf(b, "<pre><code class=\"language-xojo\">%s</code></pre>\n\n", htmlEscape(source))
	b.WriteString("</details>\n\n")
}

// looksLikeType reports whether s is a plausible type name worth linking.
// Requires an initial uppercase letter (PascalCase convention for Xojo types).
func looksLikeType(s string) bool {
	if s == "" {
		return false
	}
	r := s[0]
	return r >= 'A' && r <= 'Z'
}

func quoteIfNeeded(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return `""`
	}
	// If it contains spaces and isn't quoted, wrap it.
	if !(strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) &&
		(strings.ContainsAny(s, " \t") || strings.ContainsAny(s, "(),;")) {
		return `"` + s + `"`
	}
	return s
}
