# xojo-docgen

A Go program that parses Xojo projects and emits per-project API documentation as Markdown, then renders standalone MkDocs Material sites with an editorial Xojo theme.


## What it does

1. **Discovers** every `*.xojo_project` under a root folder.
2. **Parses** each project's `.xojo_project` manifest + `.xojo_code` / `.xojo_window` / `.xojo_menu` / `.xojo_toolbar` files.
3. **Extracts** a structured model: classes, modules, interfaces, pages, and their members (methods, properties, computed properties, constants, enums, delegates, event definitions, event handlers) — using the `#tag` layer only to *structure*, and rendering the real VB/Xojo code as the display focus.
4. **Links** type references to the official Xojo documentation via the shipped `objects.inv` Sphinx inventory (1,400+ API pages).
5. **Copies** a complete template, generates its primary-color palette, and emits Markdown + a per-project `mkdocs.yml` into `docs/api/<slug>/`.
6. `mkdocs build` then renders each project into a **standalone, deploy-ready static site** at `docs/api-published/<slug>/`.

## Build & run

```bash
# Build the binary
cd tools/docgen
go build -o xojo-docgen .

# Generate Markdown for all sample projects (under tools/sample_project/)
./xojo-docgen -root ../../tools/sample_project -out ../../docs/api -v

# Or one project
./xojo-docgen -single ../../tools/sample_project/ee_web/EEWeb.xojo_project -out ../../docs/api -v

# Omit Xojo project Folder items and their complete ParentID subtrees
./xojo-docgen -single "../../Long Pepper.xojo_project" \
  -exclude-folder "dependencies,vendor" -out ../../docs/api -v

# Use another project color (strict decimal RGB)
./xojo-docgen -root ../sample_project -out ../../docs/api \
  -primary-color "122,31,43" -v

# Use a complete project-specific template (single-project mode only)
./xojo-docgen -single "../../Long Pepper.xojo_project" \
  -template-dir ../../docs/templates/long-pepper -out ../../docs/api -v

# Then build all sites
cd ../..
make docs
```

## Flags

| Flag | Default | Purpose |
|---|---|---|
| `-root <dir>` | `tools/sample_project` | Root to scan for `*.xojo_project` (recursive). Each becomes a separate doc set. |
| `-single <file>` | — | Process just one `.xojo_project`. |
| `-exclude-folder <names>` | — | Omit comma-separated Xojo `Folder` names (case-insensitive) and their complete ParentID subtrees. |
| `-template-dir <dir>` | `templates/default` | Complete template for `-single`. The default is resolved beside the executable, then from the working tree. |
| `-primary-color <R,G,B>` | `11,99,56` | Primary color in decimal RGB. Generates light, dark, soft, border, and contrast-safe accent variants for every processed project. |
| `-out <dir>` | `docs/api` | Output dir for generated Markdown. |
| `-docs <path>` | auto-detect | Path to the Xojo `Documentation` dir (for `objects.inv`). |
| `-no-links` | false | Disable external links to official Xojo docs. |
| `-include-private` | true | Include private members (collapsed under `<details>`). |
| `-publish-prep` | false | Write `.nojekyll` into each `docs/api-published/<slug>/` for GitHub Pages readiness. |
| `-v` | false | Verbose output. |

## Project layout

```
tools/docgen/
├── go.mod                  Go module
├── main.go                 CLI: discover projects, loop, render
├── manifest.go             parse .xojo_project (config + item tree)
├── manifest_test.go        manifest hierarchy exclusion tests
├── parser.go               #tag / Begin-End two-mode scanner (core)
├── inline.go               quote-aware comma-separated property parser
├── signature.go            parse Sub/Function lines + property declarations
├── model.go                data model (Project, Container, Member types)
├── docs.go                 documentation-extraction precedence helpers
├── linkmap.go              parse objects.inv → Name→URL link map
├── render_markdown.go      emit per-project Markdown (code-focused, type-linked)
├── emit_mkdocs.go          emit per-project mkdocs.yml
├── template.go             resolve, validate, and copy complete templates
├── primary_color.go        parse RGB and generate the derived CSS palette
├── templates/default/      built-in editorial MkDocs template
│   ├── mkdocs.base.yml
│   ├── assets/
│   ├── javascripts/
│   └── stylesheets/
└── README.md               this file
```

## Architecture notes

- **Two source layers.** Xojo text files have `#tag` structural blocks (hidden) and the VB/Xojo code (displayed). The extractor uses `#tag` only to find entities; what it renders is the real code.
- **One parser covers everything.** The `#tag` + `Begin/End` grammar is uniform across `.xojo_code` / `.xojo_window` / `.xojo_menu` / `.xojo_toolbar` and across all project types (Console, Desktop, iOS, Mobile, Web).
- **External links.** The `objects.inv` shipped with the Xojo IDE maps canonical PascalCase names (`WebButton`, `Integer`, `SQLiteDatabase`) to their official doc URLs. The extractor links type tokens that follow `As` in signatures.
- **Per-project standalone sites.** Each `docs/api-published/<slug>/` is a complete static site with its own `index.html`, search, and `.nojekyll` — independently deployable to GitHub Pages or any static host.
- **Hierarchy exclusions.** `-exclude-folder` matches Xojo `Folder` item names case-insensitively and follows ParentID relationships rather than filesystem paths. Each generated project directory is replaced during regeneration so stale pages from an excluded subtree cannot remain.
- **Generated output is replaceable.** Every run removes and recreates `docs/api/<slug>/`. Do not store hand-written files in that generated project directory.
- **Templates are source assets.** The default theme is ordinary files under `templates/default`, not hard-coded Go. A custom template must contain the same required paths, including `stylesheets/primary-color.css`; the copied palette file is regenerated without changing the source template.
- **One color input, coherent variants.** `-primary-color` accepts only `R,G,B`. The generator mixes the base with white/black for its ramp and adjusts link accents until they meet a 4.5:1 contrast target on the light and dark surfaces.

## Known limitations

- Standalone language `Enum` and ComputedProperty `Setter` are grammatically supported but not exercised by the sample fixtures.
- The link map requires the Xojo IDE's `objects.inv`; use `-no-links` if absent.
- MkDocs must be installed separately (see `docs/setup-guide.md`).
