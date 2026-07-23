# Third-Party Notices

This document records the third-party work used to build, compile, test, or
publish `xojo-docgen`. Versions are the versions tested by this repository.
The linked upstream license remains authoritative.

## Current documentation build toolchain

These packages run while documentation is built. They are not the visible
published theme and are not copied into generated sites.

| Project | Version | Copyright / authorship | License | Use |
|---|---:|---|---|---|
| [MkDocs](https://www.mkdocs.org/) | 1.6.1 | Copyright © 2014-present Tom Christie and contributors | BSD-2-Clause | Markdown rendering and static-site build pipeline |
| [Python-Markdown](https://python-markdown.github.io/) | resolved by MkDocs | Python Markdown Project; Yuri Takhteyev; Manfred Stienstra | BSD-3-Clause | Markdown parser used by MkDocs |
| [Jinja](https://jinja.palletsprojects.com/) | resolved by MkDocs | Copyright 2007 Pallets and contributors | BSD-3-Clause | Template engine used by MkDocs |
| [mkdocs-literate-nav](https://github.com/oprypin/mkdocs-literate-nav) | 0.6.3 | Copyright © 2020 Oleh Prypin | MIT | Navigation generated from `SUMMARY.md` |
| [PyMdown Extensions](https://facelessuser.github.io/pymdown-extensions/) | 11.0.1 | Copyright © 2014-2025 Isaac Muse and contributors | MIT, with component notices in its license | Markdown details, highlighting metadata, tabs, task lists, and fenced blocks |

The directly installed versions are pinned in
[`requirements-docs.txt`](requirements-docs.txt). Python-Markdown and Jinja
are dependencies of MkDocs and resolve through that environment.

### Resolved Python build closure

A clean installation from `requirements-docs.txt` on 2026-07-24 resolved the
following additional build-time packages. They remain the work of their
respective authors and contributors. They are not copied into published
documentation sites.

| Package | Resolved version | Primary author / steward | License |
|---|---:|---|---|
| [Click](https://click.palletsprojects.com/) | 8.4.2 | Pallets | BSD-3-Clause |
| [ghp-import](https://github.com/c-w/ghp-import) | 2.1.0 | Paul Joseph Davis and contributors | Apache-2.0 |
| [MarkupSafe](https://markupsafe.palletsprojects.com/) | 3.0.3 | Pallets | BSD-3-Clause |
| [mergedeep](https://github.com/clarketm/mergedeep) | 1.3.4 | Travis Clarke | MIT |
| [mkdocs-get-deps](https://github.com/mkdocs/get-deps) | 0.2.2 | Oleh Prypin and contributors | MIT |
| [packaging](https://packaging.pypa.io/) | 26.2 | Python Packaging Authority and contributors | Apache-2.0 OR BSD-2-Clause |
| [pathspec](https://github.com/cpburnz/python-pathspec) | 1.1.1 | Caleb P. Burns and contributors | MPL-2.0 |
| [platformdirs](https://platformdirs.readthedocs.io/) | 4.11.0 | Bernát Gábor and contributors | MIT |
| [ProperDocs](https://properdocs.org/) | 1.6.7 | Tom Christie and contributors | BSD-2-Clause |
| [python-dateutil](https://dateutil.readthedocs.io/) | 2.9.0.post0 | Gustavo Niemeyer and contributors | BSD-3-Clause OR Apache-2.0 |
| [PyYAML](https://pyyaml.org/) | 6.0.3 | Kirill Simonov and contributors | MIT |
| [pyyaml-env-tag](https://github.com/waylan/pyyaml-env-tag) | 1.1 | Waylan Limberg | MIT |
| [six](https://github.com/benjaminp/six) | 1.17.0 | Benjamin Peterson | MIT |
| [watchdog](https://github.com/gorakhargosh/watchdog/) | 6.0.0 | Mickaël Schoentgen and contributors | Apache-2.0 |

ProperDocs is currently installed transitively by mkdocs-literate-nav. DocGen
continues to invoke the `mkdocs` command and does not select ProperDocs as its
publisher.

## Assets shipped in every published site

| Project | Copyright / authorship | License | Use |
|---|---|---|---|
| [Prism.js](https://prismjs.com/) | Copyright © 2012 Lea Verou and contributors | MIT | Client-side syntax highlighting |
| [Xojo Syntax Highlight for Web](https://github.com/jedt3d/xojo-syntax-highlight-for-web) | Worajedt Sitthidumrong | MIT | Xojo language grammar for Prism.js |
| [AntV X6](https://x6.antv.antgroup.com/) | Copyright © 2021-2023 Alipay, Inc. and contributors | MIT | Interactive ER-diagram canvas |
| [Dagre](https://github.com/dagrejs/dagre) | Copyright © 2012-2014 Chris Pettitt and contributors | MIT | Deterministic ER-diagram layout |

License texts for the vendored assets are distributed under
`docgen/templates/default/vendor/` and are copied into each generated site.

## Modules compiled into xojo-docgen

The list below is the production module closure reported by `go list -deps`
for the current Go module. Test-only modules are not listed.

| Go module | Version | Copyright / authorship | License |
|---|---:|---|---|
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | 1.29.10 | Copyright © 2017 The SQLite Authors | BSD-3-Clause |
| [modernc.org/libc](https://pkg.go.dev/modernc.org/libc) | 1.49.3 | Copyright © 2017 The Libc Authors; incorporated Go portions retain their notices | BSD-3-Clause |
| [modernc.org/mathutil](https://pkg.go.dev/modernc.org/mathutil) | 1.6.0 | Copyright © 2014 The mathutil Authors | BSD-3-Clause |
| [modernc.org/memory](https://pkg.go.dev/modernc.org/memory) | 1.8.0 | Copyright © 2017 The Memory Authors; incorporated Go and mmap portions retain their notices | BSD-3-Clause |
| [github.com/dustin/go-humanize](https://github.com/dustin/go-humanize) | 1.0.1 | Copyright © 2005-2008 Dustin Sallings | MIT |
| [github.com/google/uuid](https://github.com/google/uuid) | 1.6.0 | Copyright © 2009, 2014 Google Inc. | BSD-3-Clause |
| [github.com/mattn/go-isatty](https://github.com/mattn/go-isatty) | 0.0.20 | Copyright © Yasuhiro Matsumoto | MIT |
| [github.com/ncruces/go-strftime](https://github.com/ncruces/go-strftime) | 0.1.9 | Copyright © 2022 Nuno Cruces | MIT |
| [github.com/remyoudompheng/bigfft](https://github.com/remyoudompheng/bigfft) | 2023-01-29 revision | Copyright © 2012 The Go Authors | BSD-3-Clause |
| [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys) | 0.19.0 | Copyright © 2009 The Go Authors | BSD-3-Clause |

The complete license text for each Go module is included in that module's
source distribution and Go module cache. Binary distributors should include
those texts with their release artifacts.

## Xojo sample projects and documentation data

The Eddie's Electronics sample projects under `sample_project/` are copyright
Xojo, Inc. They are test fixtures and are not covered by DocGen's MIT license.
See [`sample_project/NOTICE`](sample_project/NOTICE).

The `objects.inv` link map is read from a user's Xojo installation. It is not
stored in this repository or copied into generated sites. Xojo names,
documentation, and the Xojo-inspired default color remain properties of Xojo,
Inc.

## Historical acknowledgment: Material for MkDocs

The initial DocGen publisher used
[Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) by Martin
Donath and contributors. Material for MkDocs is MIT-licensed, copyright
© 2016-2025 Martin Donath.

Material helped establish the first working documentation pipeline and visual
prototype. Commit `4285b2f` replaced it with DocGen's standalone Landmark
template. The current repository does not import, inherit, vendor, load, or
require Material templates, DOM components, CSS, JavaScript, icons, palette,
search index, or Python package.

This historical credit does not describe a current software dependency and
does not imply that the Landmark template contains Material code.
