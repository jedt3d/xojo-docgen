# Installing xojo-docgen

This guide installs the Go extractor and the Python documentation build
toolchain. The tested dependency versions are pinned in
[`requirements-docs.txt`](requirements-docs.txt).

## Requirements

- Go 1.21 or newer
- Python 3.10 or newer
- Xojo 2025r3 or newer only when links to the official Xojo documentation are
  required; use `-no-links` when Xojo is unavailable

MkDocs is the Markdown renderer and static-site build pipeline. The published
interface is the standalone Landmark template under
`docgen/templates/default/`. It does not use the Material for MkDocs theme,
DOM, CSS, JavaScript, icons, search index, or runtime.

## Install with pipx

`pipx` keeps the documentation toolchain separate from the system Python:

```bash
pipx install "mkdocs==1.6.1"
pipx inject mkdocs \
  "mkdocs-literate-nav==0.6.3" \
  "pymdown-extensions==11.0.1"
```

PyMdown Extensions is a direct requirement. The default configuration uses
`details`, `superfences`, `highlight`, `inlinehilite`, `tabbed`, `tasklist`,
and `tilde`.

On systems where pipx records a `uv` backend that is temporarily incompatible
with the installed `uv`, append `--backend pip` to the `install` and `inject`
commands.

## Install in a virtual environment

```bash
python3 -m venv .venv-docs
source .venv-docs/bin/activate
python -m pip install -r requirements-docs.txt
```

On Windows PowerShell, activate with:

```powershell
.\.venv-docs\Scripts\Activate.ps1
```

## Build the extractor

```bash
cd docgen
go build -o xojo-docgen .
go test ./...
```

## Generate and publish

Generate one Xojo text project:

```bash
./xojo-docgen \
  -single /path/to/App.xojo_project \
  -out /path/to/docs/api \
  -publish-prep
```

Build the generated site:

```bash
mkdocs build --strict -f /path/to/docs/api/app/mkdocs.yml
```

The result is written beside `api/` under `api-published/app/`. It is a
standalone static site suitable for GitHub Pages or another static host.

DocGen requires `-single` or `-root`; the repository does not distribute or
implicitly select example Xojo projects. The tracked sites under `docs/` are
generated evaluation snapshots only.

The default Landmark template is stored at `docgen/templates/default/`. A
binary run from the cloned `docgen/` directory finds it automatically. When
distributing the executable separately, copy that directory to
`templates/default/` beside the executable or provide a complete template with
`-template-dir` in single-project mode.

For the complete generation, customization, preview, and deployment workflow,
see [`docs/setup-guide.md`](docs/setup-guide.md).

## Official Xojo links

DocGen auto-detects the `objects.inv` file shipped with the Xojo IDE. Override
the location with `-docs /path/to/Documentation`, or disable official links
with `-no-links`.

## Migrating from the original Material-based toolchain

The initial DocGen implementation used Material for MkDocs. Since commit
`4285b2f`, it has used its own Landmark publishing template. Material is not a
current dependency.

To replace an older pipx environment completely:

```bash
pipx uninstall mkdocs
pipx install "mkdocs==1.6.1"
pipx inject mkdocs \
  "mkdocs-literate-nav==0.6.3" \
  "pymdown-extensions==11.0.1"
```

Verify that the environment contains the three packages above and does not
contain `mkdocs-material` or `mkdocs-material-extensions`:

```bash
pipx list
```

## Troubleshooting

### `externally-managed-environment`

Do not use `--break-system-packages`. Use pipx or a virtual environment.

### `No module named pymdownx`

Install PyMdown Extensions into the same environment as MkDocs:

```bash
pipx inject mkdocs "pymdown-extensions==11.0.1"
```

### `literate-nav` is not installed

```bash
pipx inject mkdocs "mkdocs-literate-nav==0.6.3"
```

### `mkdocs.yml` does not exist

Run the generator first, then pass the generated per-project configuration to
MkDocs with `-f`.

## Attribution

See [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md) for the current build,
runtime, vendored, and compiled dependencies and for the historical Material
for MkDocs acknowledgment.
