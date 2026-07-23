# Xojo DocGen Setup and Publishing Guide

This guide covers the current standalone Landmark release of `xojo-docgen`.
Start with [`../INSTALLATION.md`](../INSTALLATION.md) for the tested Go and
Python versions.

## 1. Build the command-line tool

From the repository root:

```bash
cd docgen
go build -o xojo-docgen .
go test ./...
cd ..
```

The executable resolves the default template from
`docgen/templates/default/` when run in the repository. For a packaged binary,
distribute the template as `templates/default/` beside the executable.

## 2. Choose an output location

DocGen creates one directory per Xojo project. Use a private build directory
when experimenting so the release examples under `docs/` remain unchanged:

```bash
export XOJO_DOCS_OUTPUT="/absolute/path/to/output/api"
```

The published directory will be created beside `api/` as
`api-published/<project-slug>/`.

## 3. Generate one Xojo project

Pass a Xojo text-project manifest explicitly:

```bash
docgen/xojo-docgen \
  -single "/absolute/path/to/MyApp.xojo_project" \
  -out "$XOJO_DOCS_OUTPUT" \
  -v
```

DocGen requires either `-single` or `-root`; it does not depend on bundled
example projects.

### Exclude project folders

`-exclude-folder` accepts a comma-separated list of Xojo `Folder` item names.
Matching is case-insensitive and follows the manifest's `ParentID` hierarchy:

```bash
docgen/xojo-docgen \
  -single "/absolute/path/to/MyApp.xojo_project" \
  -exclude-folder "dependencies,vendor" \
  -out "$XOJO_DOCS_OUTPUT"
```

### Choose the primary color

Provide three decimal RGB channels from 0 through 255:

```bash
docgen/xojo-docgen \
  -single "/absolute/path/to/MyApp.xojo_project" \
  -primary-color "122,31,43" \
  -out "$XOJO_DOCS_OUTPUT"
```

DocGen derives the light, dark, soft, border, and contrast-safe accent variants
and writes them into the generated copy of
`stylesheets/primary-color.css`. The source template is unchanged.

### Document SQLite schemas

`-database` is repeatable and restricted to `-single`. Relative paths are
resolved from the selected `.xojo_project`:

```bash
docgen/xojo-docgen \
  -single "/absolute/path/to/MyApp.xojo_project" \
  -database "data/app.sqlite" \
  -out "$XOJO_DOCS_OUTPUT"
```

The database is opened read-only and query-only. DocGen extracts schema
metadata, not application rows. See
[`../docgen/DATABASE_DOCUMENTATION.md`](../docgen/DATABASE_DOCUMENTATION.md).

### Official Xojo API links

DocGen auto-detects the `objects.inv` inventory installed with the Xojo IDE.
Override it with `-docs /path/to/Documentation`, or generate without external
links:

```bash
docgen/xojo-docgen \
  -single "/absolute/path/to/MyApp.xojo_project" \
  -no-links \
  -out "$XOJO_DOCS_OUTPUT"
```

## 4. Generate multiple projects

Use `-root` to scan recursively for `.xojo_project` manifests:

```bash
docgen/xojo-docgen \
  -root "/absolute/path/to/xojo-projects" \
  -out "$XOJO_DOCS_OUTPUT" \
  -v
```

Each manifest produces an independent `<slug>/` directory. Project-specific
templates and SQLite inputs require `-single` so they cannot be associated with
the wrong project.

## 5. Use a custom Landmark template

Copy the complete default template outside the generated output:

```bash
cp -R docgen/templates/default /absolute/path/to/my-template
```

Then generate one project:

```bash
docgen/xojo-docgen \
  -single "/absolute/path/to/MyApp.xojo_project" \
  -template-dir "/absolute/path/to/my-template" \
  -out "$XOJO_DOCS_OUTPUT"
```

The template must retain all required HTML, hook, JavaScript, stylesheet,
featured-image, database, and vendored-license files. DocGen validates the
directory before replacing any generated project output.

## 6. Build the published site

The generated configuration records the correct sibling `api-published`
destination:

```bash
mkdocs build --strict \
  -f "$XOJO_DOCS_OUTPUT/myapp/mkdocs.yml"
```

The result is a standalone static site under
`/absolute/path/to/output/api-published/myapp/`. The Landmark HTML, navigation,
search, syntax highlighting, theme switching, database viewer, and assets are
local to that site.

## 7. Preview locally

Serve one published project:

```bash
python3 -m http.server 8000 \
  --directory "/absolute/path/to/output/api-published/myapp"
```

Or serve every published project with hostname routing:

```bash
python3 docgen/serve.py \
  --root "/absolute/path/to/output/api-published" \
  --port 8910
```

Open `http://myapp.lvh.me:8910/`.

## 8. Publish

Upload the contents of one `api-published/<slug>/` directory to GitHub Pages,
Netlify, Vercel, S3, nginx, or another static host. No Xojo runtime, Go binary,
Python process, database, or original project source is required by the
published site.

## 9. Generated examples in this repository

[`README.md`](README.md) describes the six tracked evaluation snapshots under
`api/` and `api-published/`. They show current output across Xojo targets but
are never read by the generator. The original example projects are not
distributed.

## Troubleshooting

### `provide -root <dir> or -single <project.xojo_project>`

Supply the project manifest or a directory containing manifests. There is no
implicit example-project fallback.

### `default template directory not found`

Run the compiled binary from the cloned `docgen/` directory, keep
`templates/default/` beside a distributed executable, or pass a complete
template with `-template-dir` in single-project mode.

### `No module named pymdownx`

Install the pinned documentation dependencies from
[`../requirements-docs.txt`](../requirements-docs.txt) in the same environment
that provides `mkdocs`.

### `objects.inv` not found

Pass the Xojo Documentation directory with `-docs`, or use `-no-links`.

### `mkdocs.yml` does not exist

Run DocGen first and use the configuration inside the generated project slug.

## Architecture and attribution

MkDocs renders Markdown and runs the build pipeline. The visible DOM, styling,
navigation, search, and client behavior come from DocGen's standalone Landmark
template (`theme.name: null`). Material for MkDocs is a historical
acknowledgment, not a current dependency.

See [`../THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md) for current,
vendored, compiled, and historical dependency notices.
