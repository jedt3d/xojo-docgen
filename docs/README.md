# Documentation and Generated Examples

This directory contains the documentation shipped with `xojo-docgen` and the
output snapshots used to review the current release.

## Contents

| Path | Purpose | Maintained by |
|---|---|---|
| `setup-guide.md` | Generate, build, preview, and publish documentation | Hand-written |
| `naming-guide.md` | Xojo naming reference used while reviewing generated APIs | Hand-written |
| `api/<slug>/` | Generated Markdown, Landmark assets, manifests, and MkDocs configuration | `xojo-docgen` |
| `api-published/<slug>/` | Ready-to-host static Landmark sites | `mkdocs build` |

Do not hand-edit files below `api/` or `api-published/`. Regeneration replaces a
project's complete generated directory.

## Included evaluation results

The snapshots were produced while testing DocGen against these Xojo example
applications:

| Slug | Example | Target represented |
|---|---|---|
| `sendingemail` | SendingEmail | Console |
| `eedesktop` | EEDesktop | Desktop |
| `eeios` | EEiOS | iOS |
| `eeandroid` | EEAndroid | Mobile/Android |
| `eeweb` | EEWeb | Web |
| `eewebservices` | EEWebServices | Web service |

These directories are expected output, not generator inputs. The original Xojo
projects are not distributed in this repository. To reproduce the snapshots,
obtain the examples through an appropriately licensed Xojo installation and
run the current release against those local project files.

Example-project names and generated material derived from those projects remain
attributable to Xojo, Inc. and are outside DocGen's MIT license. See
[`../THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md).

## View a published example

Any static HTTP server can serve one project:

```bash
python3 -m http.server 8000 --directory docs/api-published/eedesktop
```

Then open <http://127.0.0.1:8000/>.

To serve every snapshot through project-specific `lvh.me` hostnames:

```bash
python3 docgen/serve.py --root docs/api-published --port 8910
```

For example:

- <http://eedesktop.lvh.me:8910/>
- <http://eeweb.lvh.me:8910/>

See [`setup-guide.md`](setup-guide.md) to generate documentation for another
Xojo project.
