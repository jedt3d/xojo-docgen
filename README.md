# xojo-docgen

**Generate API documentation from Xojo projects — Go → Markdown → MkDocs Material.**

`xojo-docgen` parses Xojo text-project files (`.xojo_project` + `.xojo_code` / `.xojo_window` / `.xojo_menu` / `.xojo_toolbar`), extracts every documentable entity, and emits clean Markdown — then renders each project into a standalone, deploy-ready static site themed in Xojo green.

- **Code as the display focus.** The `#tag` structural layer is used only to extract entities; what you see in the docs is the real VB/Xojo source (signatures + collapsible full method bodies).
- **Per-project sites.** Each Xojo project becomes its own independent MkDocs Material site, ready to publish to GitHub Pages or any static host.
- **Multi-project.** Point the tool at a folder of projects; it generates a separate doc set for each. Built and tested against all five Xojo project types (Console, Desktop, iOS, Mobile/Android, Web).
- **Official-docs linking.** Type references (`As WebButton`, `As Integer`, `Inherits SQLiteDatabase`) auto-link to the official Xojo documentation via the IDE's shipped `objects.inv` inventory.
- **Syntax highlighting.** Full Xojo grammar via Prism.js, with a green-tuned token palette that harmonizes with the brand.
- **Source review.** Every method body sits in a collapsible block with a fullscreen modal for long code review.

---

## The story

Xojo is a cross-platform development tool for building Desktop, Web, and Mobile apps. Its text-based "Xojo Project" format saves each project item as a separate diff-friendly file — perfect for Git, but there was no good way to generate API documentation from it like you can with Javadoc, Sphinx, or YARD.

`xojodoc` (a community tool) existed but was unmaintained and written in Xojo itself, making it hard to run in CI. `xojo-docgen` takes a different approach: a purpose-built Go extractor that parses the `#tag` format directly, paired with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/) for publishing. The result is fast, self-contained, and produces genuinely beautiful docs.

---

## Quick start

```bash
# Prerequisites: Go 1.21+, Python 3.10+, MkDocs + Material (see docs/setup-guide.md)

# Build the extractor
cd docgen
go build -o xojo-docgen .

# Generate Markdown for all sample projects
./xojo-docgen -root ../sample_project -out ../../docs/api -v

# Render each project into a standalone site
cd ../..
make docs

# Preview one project
make docs-serve PROJECT=eeweb    # http://127.0.0.1:8000

# Preview all projects on their own local domains
make docs-serve-all              # http://eeweb.lvh.me:8910/, etc.
```

## How it works

```
.xojo_project + .xojo_code ──► xojo-docgen (Go) ──► docs/api/<slug>/*.md
                                                         │
                                                         ▼
                                          mkdocs build (per project)
                                                         │
                                                         ▼
                                     docs/api-published/<slug>/  ← standalone site
```

Each published site is a complete static site — its own `index.html`, search index, `.nojekyll`, and assets. Deploy one to GitHub Pages or drop it on any static host.

---

## Open source projects used

`xojo-docgen` stands on the shoulders of excellent open source work. Full credit to these projects and their authors:

| Project | Author | License | Purpose | URL |
|---|---|---|---|---|
| **Go** | The Go Authors | BSD-3-Clause | The extractor language | [go.dev](https://go.dev) |
| **MkDocs** | Tom Christie | BSD-2-Clause | Static site generator for the docs | [mkdocs.org](https://www.mkdocs.org) |
| **MkDocs Material** | Martin Donath | MIT | The documentation theme | [squidfunk.github.io/mkdocs-material](https://squidfunk.github.io/mkdocs-material/) |
| **mkdocs-literate-nav** by Tim Schwenke | MIT | Auto-builds the nav from the file tree | [github.com/oprypin/mkdocs-literate-nav](https://github.com/oprypin/mkdocs-literate-nav) |
| **Prism.js** | Lea Verou & James DiGioia | MIT | Client-side syntax highlighting | [prismjs.com](https://prismjs.com) |
| **Xojo Prism grammar** | Worajedt Sitthidumrong | MIT | The Xojo language definition for Prism | [github.com/jedt3d/xojo-syntax-highlight-for-web](https://github.com/jedt3d/xojo-syntax-highlight-for-web) |

The Xojo brand color (`#87B946`) and documentation link map (`objects.inv`) are properties of Xojo, Inc.

---

## Licensing & copyright

### The `xojo-docgen` tool (`docgen/`)

Copyright © 2026 Worajedt Sitthidumrong. Licensed under the **MIT License** — see [LICENSE](LICENSE).

### The sample projects (`sample_project/`)

The "Eddie's Electronics" sample applications under `sample_project/` are **© Xojo, Inc.** They are included only as test fixtures and are **not** covered by the MIT license. See [sample_project/NOTICE](sample_project/NOTICE) and [xojo.com/license](https://www.xojo.com/license/).

### Vendored third-party assets (`docgen/assets/`)

- `prism.js` — © PrismJS contributors, MIT License ([prismjs.com](https://prismjs.com))
- `xojo.prism.js` — © Worajedt Sitthidumrong, MIT License

---

## Project layout

```
xojo-docgen/
├── LICENSE                    MIT — the docgen tool
├── README.md                  this file
├── docgen/                    the Go extractor
│   ├── go.mod
│   ├── main.go                CLI: discover projects, loop, render
│   ├── manifest.go            parse .xojo_project
│   ├── parser.go              #tag / Begin-End two-mode scanner (core)
│   ├── inline.go              quote-aware property parser
│   ├── signature.go           parse Sub/Function + property declarations
│   ├── model.go               data model
│   ├── docs.go                documentation extraction
│   ├── linkmap.go             objects.inv → official-docs link map
│   ├── featured.go            green placeholder PNG generator
│   ├── render_markdown.go     Markdown rendering (code-focused, type-linked)
│   ├── emit_mkdocs.go         per-project mkdocs.yml emitter
│   ├── theme.go               embedded CSS/JS (go:embed)
│   ├── extra.css              Xojo green theme stylesheet
│   ├── mkdocs.base.yml        shared MkDocs config
│   ├── assets/                vendored Prism.js + Xojo grammar + modal JS
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
- **MkDocs Material:** [squidfunk.github.io/mkdocs-material](https://squidfunk.github.io/mkdocs-material/)
- **Xojo syntax highlight:** [github.com/jedt3d/xojo-syntax-highlight-for-web](https://github.com/jedt3d/xojo-syntax-highlight-for-web)

## Contributing

