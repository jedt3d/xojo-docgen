# xojo-docgen

A Go program that parses Xojo projects and emits per-project API documentation as Markdown, then publishes it through the canonical EEWeb editorial reader.

> Lives under `tools/`, which is **git-ignored** from the Long Pepper repo.

Install the Go and Python build toolchains from
[`../INSTALLATION.md`](../INSTALLATION.md). MkDocs, mkdocs-literate-nav, and
PyMdown Extensions are the current Python requirements.

## What it does

1. **Discovers** every `*.xojo_project` under a root folder.
2. **Parses** each project's `.xojo_project` manifest + `.xojo_code` / `.xojo_window` / `.xojo_menu` / `.xojo_toolbar` files.
3. **Extracts** a structured model: classes, modules, interfaces, pages, and their members (methods, properties, computed properties, constants, enums, delegates, event definitions, event handlers) — using the `#tag` layer only to *structure*, and rendering the real VB/Xojo code as the display focus.
4. **Links** type references to the official Xojo documentation via the shipped `objects.inv` Sphinx inventory (1,400+ API pages).
5. **Optionally inspects** explicitly selected SQLite files in read-only mode and emits a data dictionary plus ER relationship model.
6. **Copies** the EEWeb editorial template, emits its project/entity/database manifests, generates its primary-color palette, and writes Markdown source into `docs/api/<slug>/content/` plus a per-project `mkdocs.yml`.
7. `mkdocs build` renders Markdown, runs the Landmark payload hook, and publishes a **standalone, deploy-ready static site** at `docs/api-published/<slug>/`.

## Build & run

```bash
# Build the binary
cd tools/docgen
go build -o xojo-docgen .

# Generate Markdown for all sample projects (under tools/sample_project/)
./xojo-docgen -root ../../tools/sample_project -out ../../docs/api -v

# Or one project
./xojo-docgen -single ../../tools/sample_project/ee_web/EEWeb.xojo_project -out ../../docs/api -v

# Add a project-relative SQLite database schema (repeatable)
./xojo-docgen -single "../../dependencies/XjMVVM/mvvm.xojo_project" \
  -database data/notes.sqlite -out ../../docs/api -v

# Omit Xojo project Folder items and their complete ParentID subtrees
./xojo-docgen -single "../../Long Pepper.xojo_project" \
  -exclude-folder "dependencies,vendor" -out ../../docs/api -v

# Use another project color (strict decimal RGB)
./xojo-docgen -root ../sample_project -out ../../docs/api \
  -primary-color "122,31,43" -v

# Use a complete project-specific template (single-project mode only)
./xojo-docgen -single "../../Long Pepper.xojo_project" \
  -template-dir /path/to/custom-template -out ../../docs/api -v

# Then build all sites
cd ../..
make docs
```

## Flags

| Flag | Default | Purpose |
|---|---|---|
| `-root <dir>` | `tools/sample_project` | Root to scan for `*.xojo_project` (recursive). Each becomes a separate doc set. |
| `-single <file>` | — | Process just one `.xojo_project`. |
| `-database <file>` | — | SQLite database to document. Repeatable, resolved relative to the `.xojo_project`, and restricted to `-single`. |
| `-exclude-folder <names>` | — | Omit comma-separated Xojo `Folder` names (case-insensitive) and their complete ParentID subtrees. |
| `-template-dir <dir>` | `templates/default` | Complete template for `-single`. The default is resolved beside the executable, then from the working tree. |
| `-primary-color <R,G,B>` | `11,99,56` | Primary color in decimal RGB. Generates light, dark, soft, border, and contrast-safe accent variants for every processed project. |
| `-out <dir>` | `docs/api` | Output dir for generated Markdown. |
| `-docs <path>` | auto-detect | Path to the Xojo `Documentation` dir (for `objects.inv`). |
| `-no-links` | false | Disable external links to official Xojo docs. |
| `-include-private` | true | Include private members in the generated documentation. |
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
├── editorial_manifest.go   project/entity payload consumed by the reader
├── database.go             read-only SQLite schema inspection + relation model
├── database_render.go      database payload and searchable Markdown
├── database_test.go        schema extraction and inference tests
├── DATABASE_DOCUMENTATION.md  database feature contract and research
├── templates/default/      canonical EEWeb editorial publishing template
│   ├── mkdocs.base.yml
│   ├── assets/
│   ├── hooks/
│   ├── javascripts/
│   ├── overrides/
│   └── stylesheets/
└── README.md               this file
```

## Architecture notes

- **Two source layers.** Xojo text files have `#tag` structural blocks (hidden) and the VB/Xojo code (displayed). The extractor uses `#tag` only to find entities; what it renders is the real code.
- **One parser covers everything.** The `#tag` + `Begin/End` grammar is uniform across `.xojo_code` / `.xojo_window` / `.xojo_menu` / `.xojo_toolbar` and across all project types (Console, Desktop, iOS, Mobile, Web).
- **External links.** The `objects.inv` shipped with the Xojo IDE maps canonical PascalCase names (`WebButton`, `Integer`, `SQLiteDatabase`) to their official doc URLs. The extractor links type tokens that follow `As` in signatures.
- **Per-project standalone sites.** Each `docs/api-published/<slug>/` is a complete static site with its own `index.html`, generated Landmark document payload, client search, and `.nojekyll` — independently deployable to GitHub Pages or any static host.
- **Hierarchy exclusions.** `-exclude-folder` matches Xojo `Folder` item names case-insensitively and follows ParentID relationships rather than filesystem paths. Each generated project directory is replaced during regeneration so stale pages from an excluded subtree cannot remain.
- **Generated output is replaceable.** Every run removes and recreates `docs/api/<slug>/`. Do not store hand-written files in that generated project directory.
- **Templates are source assets.** The default theme is ordinary files under `templates/default`, not hard-coded Go. A custom template must contain the same required paths, including `overrides/main.html`, `hooks/editorial.py`, `javascripts/editorial.js`, `stylesheets/editorial.css`, and `stylesheets/primary-color.css`; the copied palette file is regenerated without changing the source template.
- **No Material dependency.** MkDocs uses `theme.name: null`. The Landmark override owns the complete DOM, the build hook emits `data/documents.json`, and the reader never consumes Material templates, components, bundles, or search-index HTML.
- **Historical credit is retained.** The first publisher used Material for MkDocs. It was replaced by the standalone Landmark template in commit `4285b2f`; see [`../THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md).
- **The EEWeb design is canonical.** The default reader is the approved `eeweb-docs-editorial.sjedt.chatgpt.site` interface made project-agnostic. Project names, facts, entity groups, counts, sections, search results, links, and source bodies come from generated data rather than EEWeb constants.
- **One color input, coherent variants.** `-primary-color` accepts only `R,G,B`. The generator mixes the base with white/black for its ramp and adjusts link accents until they meet a 4.5:1 contrast target on the light and dark surfaces.
- **Cache-safe regeneration.** Each generated `mkdocs.yml` includes a deterministic content fingerprint. Landmark appends it to CSS, JavaScript, database payload, and document payload URLs so a normal browser refresh loads the current generation without stale theme or API content.
- **Explicit database scope.** Database discovery is never recursive. Each `-database` path is associated with one `-single` project, opened with SQLite `mode=ro` and `PRAGMA query_only`, and represented as schema metadata only.
- **Truth before inference.** Declared SQLite foreign keys are solid ER edges. An unambiguous, type-compatible match between a non-primary source column and a target table's single primary key may produce a dashed suggested edge. This covers names such as `user_id`, `CustomerID`, `InvoiceNo`, and `ProductCode`; its evidence is included in the payload and dictionary, and it is never labeled as a constraint.
- **Large-schema LOD.** Small ER diagrams show every field on the canvas. Above 80 tables or 1,200 columns, compact topology nodes replace thousands of field rows; the inspector and dictionary retain the full detail.
- **Offline diagrams.** AntV X6 2.19.2 and Dagre 0.8.5 are pinned inside the default template, loaded only on an ER route, and copied into every published site. No CDN is required.

## Known limitations

- Standalone language `Enum` and ComputedProperty `Setter` are grammatically supported but not exercised by the sample fixtures.
- The link map requires the Xojo IDE's `objects.inv`; use `-no-links` if absent.
- The MkDocs toolchain must be installed separately; use the pinned,
  Material-free environment in [`../INSTALLATION.md`](../INSTALLATION.md).
- Database extraction currently supports SQLite 3 files. Broken view definitions are documented with an inspection diagnostic rather than aborting the entire database, while unreadable table metadata remains a build error.

Current and historical dependency credits are maintained in
[`../THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md).
