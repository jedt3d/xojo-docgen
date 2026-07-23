# xojo-docgen

**Generate API documentation from Xojo projects — Go → Markdown → EEWeb editorial reader.**

`xojo-docgen` parses Xojo text-project files (`.xojo_project` + `.xojo_code` / `.xojo_window` / `.xojo_menu` / `.xojo_toolbar`), extracts every documentable entity, and emits clean Markdown — then renders each project into a standalone, deploy-ready static site using an editable editorial Xojo template.

- **Code as the display focus.** The `#tag` structural layer is used only to extract entities; methods display their complete VB/Xojo source directly in one syntax-highlighted block.
- **Per-project sites.** Each Xojo project becomes its own independent static site, ready to publish to GitHub Pages or any static host.
- **Multi-project.** Point the tool at a folder of projects; it generates a separate doc set for each. Built and tested against all five Xojo project types (Console, Desktop, iOS, Mobile/Android, Web).
- **Official-docs linking.** Type references (`As WebButton`, `As Integer`, `Inherits SQLiteDatabase`) auto-link to the official Xojo documentation via the IDE's shipped `objects.inv` inventory.
- **Canonical EEWeb reader.** The interface approved at `eeweb-docs-editorial.sjedt.chatgpt.site` is the default publishing template: editorial overview, collapsible project rail, search, hash-addressable entity reader, sticky page contents, dark mode, and Xojo syntax highlighting.
- **Theme without recompiling.** The default reader is a normal template directory. `-template-dir` selects a complete per-project template, while `-primary-color R,G,B` generates its coordinated palette.
- **Syntax highlighting.** Full Xojo grammar via Prism.js, preserved independently of the selected primary color.
- **Source review.** Every method body remains readable in horizontally scrollable, syntax-highlighted source blocks.
- **Database documentation.** An explicit `-database` input adds a searchable SQLite data dictionary and an interactive AntV X6 ER diagram without reading application row data.

---

## The story

Xojo is a cross-platform development tool for building Desktop, Web, and Mobile apps. Its text-based "Xojo Project" format saves each project item as a separate diff-friendly file — perfect for Git, but there was no good way to generate API documentation from it like you can with Javadoc, Sphinx, or YARD.

`xojodoc` (a community tool) existed but was unmaintained and written in Xojo itself, making it hard to run in CI. `xojo-docgen` takes a different approach: a purpose-built Go extractor that parses the `#tag` format directly. MkDocs renders Markdown and runs the static build; DocGen’s EEWeb editorial reader owns the complete published DOM and interface. The result is fast, self-contained, and produces genuinely readable docs.

---

## Quick start

```bash
# Prerequisites: Go 1.21+, Python 3.10+, MkDocs,
# mkdocs-literate-nav, and PyMdown Extensions

# Build the extractor
cd docgen
go build -o xojo-docgen .

# Generate Markdown for all sample projects
./xojo-docgen -root ../sample_project -out ../../docs/api -v

# Or generate one project while omitting dependency folders from its API
./xojo-docgen -single "../../Long Pepper.xojo_project" \
  -exclude-folder "dependencies,vendor" -out ../../docs/api -v

# Add a project-relative SQLite schema (repeat -database for more than one)
./xojo-docgen -single "../../dependencies/XjMVVM/mvvm.xojo_project" \
  -database data/notes.sqlite -out ../../docs/api -v

# Generate the same editorial theme from another primary RGB value
./xojo-docgen -root ../sample_project -out ../../docs/api \
  -primary-color "122,31,43" -v

# Render each project into a standalone site
cd ../..
make docs

# Preview one project
make docs-serve PROJECT=eeweb    # http://127.0.0.1:8000

# Preview all projects on their own local domains
make docs-serve-all              # http://eeweb.lvh.me:8910/, etc.
```

`-exclude-folder` accepts a comma-separated list of Xojo `Folder` item names.
Matching is case-insensitive and follows the manifest's `ParentID` hierarchy,
so every nested item is omitted regardless of its filesystem path or declaration
order. Regeneration replaces `docs/api/<slug>/` completely to prevent stale API
pages; keep hand-written files outside that generated directory.

`-primary-color` accepts three decimal channels from 0 through 255. The default
is Xojo dark green (`11,99,56`). DocGen derives light, dark, soft, border, and
contrast-safe accent variants and writes `stylesheets/primary-color.css` into
each generated project. The source template is never modified.

`-template-dir` is intentionally limited to `-single`, preventing one
project-specific visual identity from being applied accidentally to a batch.
Copy `docgen/templates/default/` as the starting point; keep the required
directory structure, including `overrides/main.html`,
`javascripts/editorial.js`, and `stylesheets/editorial.css`. Use the
`--xojo-primary-*` CSS variables when custom styles should respond to
`-primary-color`.

`-database` is also limited to `-single`. It accepts a SQLite file path relative
to the selected `.xojo_project` (or an absolute path) and may be repeated.
DocGen opens the file in read-only/query-only mode and extracts schema metadata,
not row data. Declared foreign keys are rendered as solid ER edges. A narrowly
scoped naming heuristic may add dashed **suggested** edges when a non-primary
column uniquely matches one compatible target table and its single primary key
(for example `user_id`, `CustomerID`, `InvoiceNo`, or `ProductCode`);
suggestions are never presented as database constraints. See
[`docgen/DATABASE_DOCUMENTATION.md`](docgen/DATABASE_DOCUMENTATION.md).

## How it works

```
.xojo_project + source + optional SQLite ──► xojo-docgen ──► docs/api/<slug>/content/
                                                                    │
                                                                    ▼
                                                     mkdocs build (per project)
                                                                    │
                                                                    ▼
                                                docs/api-published/<slug>/
```

Each published site is a complete static site — its own `index.html`, Landmark document payload, client search, `.nojekyll`, and assets. Deploy one to GitHub Pages or drop it on any static host.

Installation, including the tested Python package versions, is documented in
[`INSTALLATION.md`](INSTALLATION.md).

---

## Open source projects used

`xojo-docgen` stands on the shoulders of excellent open source work. The table
below lists the directly selected build and runtime projects. The complete
compiled dependency closure, copyright notices, licenses, and historical
acknowledgments are in
[`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).

| Project | Author | License | Purpose | URL |
|---|---|---|---|---|
| **Go** | The Go Authors | BSD-3-Clause | The extractor language | [go.dev](https://go.dev) |
| **MkDocs** | Tom Christie | BSD-2-Clause | Static site generator for the docs | [mkdocs.org](https://www.mkdocs.org) |
| **mkdocs-literate-nav** | Oleh Prypin | MIT | Builds navigation from `SUMMARY.md` | [github.com/oprypin/mkdocs-literate-nav](https://github.com/oprypin/mkdocs-literate-nav) |
| **PyMdown Extensions** | Isaac Muse | MIT with component notices | Markdown extensions used by the generated source | [facelessuser.github.io/pymdown-extensions](https://facelessuser.github.io/pymdown-extensions/) |
| **Prism.js** | Lea Verou & James DiGioia | MIT | Client-side syntax highlighting | [prismjs.com](https://prismjs.com) |
| **Xojo Prism grammar** | Worajedt Sitthidumrong | MIT | The Xojo language definition for Prism | [github.com/jedt3d/xojo-syntax-highlight-for-web](https://github.com/jedt3d/xojo-syntax-highlight-for-web) |
| **modernc.org/sqlite** | modernc.org contributors | BSD-3-Clause | Pure-Go, read-only SQLite schema inspection | [pkg.go.dev/modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) |
| **AntV X6** | AntV contributors | MIT | Interactive ER diagram canvas | [x6.antv.antgroup.com](https://x6.antv.antgroup.com/) |
| **Dagre** | Dagre contributors | MIT | Directed graph layout for ER diagrams | [github.com/dagrejs/dagre](https://github.com/dagrejs/dagre) |

The default Xojo-inspired primary color and documentation link map (`objects.inv`) are properties of Xojo, Inc.

### Historical acknowledgment

The initial DocGen publisher used **Material for MkDocs**, created by Martin
Donath and contributors and licensed under MIT. It helped establish the first
working documentation pipeline and visual prototype. Commit `4285b2f` replaced
it with the standalone Landmark template. Material is not shipped, loaded, or
required by the current implementation. See the historical notice in
[`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).

---

## Licensing & copyright

### The `xojo-docgen` tool (`docgen/`)

Copyright © 2026 Worajedt Sitthidumrong. Licensed under the **MIT License** — see [LICENSE](LICENSE).

### The sample projects (`sample_project/`)

The "Eddie's Electronics" sample applications under `sample_project/` are **© Xojo, Inc.** They are included only as test fixtures and are **not** covered by the MIT license. See [sample_project/NOTICE](sample_project/NOTICE) and [xojo.com/license](https://www.xojo.com/license/).

### Vendored third-party assets (`docgen/templates/default/`)

- `prism.js` — © PrismJS contributors, MIT License ([prismjs.com](https://prismjs.com))
- `xojo.prism.js` — © Worajedt Sitthidumrong, MIT License
- `vendor/antv-x6-2.19.2.*` — © AntV contributors, MIT License
- `vendor/dagre-0.8.5.min.js` — © Dagre contributors, MIT License

---

## Project layout

```
xojo-docgen/
├── LICENSE                    MIT — the docgen tool
├── README.md                  this file
├── INSTALLATION.md            canonical setup and migration guide
├── THIRD_PARTY_NOTICES.md     current and historical dependency credit
├── requirements-docs.txt      tested Python documentation toolchain
├── docgen/                    the Go extractor
│   ├── go.mod
│   ├── main.go                CLI: discover projects, loop, render
│   ├── manifest.go            parse .xojo_project
│   ├── manifest_test.go       manifest hierarchy exclusion tests
│   ├── parser.go              #tag / Begin-End two-mode scanner (core)
│   ├── inline.go              quote-aware property parser
│   ├── signature.go           parse Sub/Function + property declarations
│   ├── model.go               data model
│   ├── docs.go                documentation extraction
│   ├── linkmap.go             objects.inv → official-docs link map
│   ├── render_markdown.go     Markdown rendering (code-focused, type-linked)
│   ├── emit_mkdocs.go         per-project mkdocs.yml emitter
│   ├── template.go            resolve, validate, and copy templates
│   ├── primary_color.go       RGB parsing + derived palette generation
│   ├── editorial_manifest.go  project/entity payload for the editorial reader
│   ├── database.go            read-only SQLite schema inspection
│   ├── database_render.go     database JSON + Markdown output
│   ├── DATABASE_DOCUMENTATION.md  database documentation design
│   ├── templates/default/     built-in editorial template
│   │   ├── mkdocs.base.yml
│   │   ├── assets/
│   │   ├── hooks/             rendered-document payload hook
│   │   ├── overrides/         complete EEWeb publishing shell
│   │   ├── javascripts/       reader runtime + database viewer + Prism grammar
│   │   └── stylesheets/       generated palette + canonical EEWeb CSS
│   └── README.md              extractor-specific docs
└── sample_project/            Xojo sample projects (© Xojo, Inc.) — test fixtures
    ├── NOTICE                 Xojo copyright notice
    ├── console_sending_email/
    ├── ee_android/
    ├── ee_desktop/
    ├── ee_ios/
    ├── ee_web/
    └── ee_webservices/
```

---

## Links

- **Repository:** [github.com/jedt3d/xojo-docgen](https://github.com/jedt3d/xojo-docgen)
- **Issues:** [github.com/jedt3d/xojo-docgen/issues](https://github.com/jedt3d/xojo-docgen/issues)
- **Xojo:** [xojo.com](https://www.xojo.com)
- **MkDocs:** [mkdocs.org](https://www.mkdocs.org/)
- **Xojo syntax highlight:** [github.com/jedt3d/xojo-syntax-highlight-for-web](https://github.com/jedt3d/xojo-syntax-highlight-for-web)
- **Installation:** [INSTALLATION.md](INSTALLATION.md)
- **Third-party notices:** [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)

## Contributing

Pull requests welcome. The extractor is pure Go; SQLite schema inspection uses
the pure-Go `modernc.org/sqlite` driver. Architecture notes are maintained in
[`docgen/README.md`](docgen/README.md), with the database contract in
[`docgen/DATABASE_DOCUMENTATION.md`](docgen/DATABASE_DOCUMENTATION.md).
