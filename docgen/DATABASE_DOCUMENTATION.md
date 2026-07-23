# Database documentation

This document defines the first database-documentation slice for
`xojo-docgen`. The implementation is intentionally SQLite-first and keeps the
published interface in the canonical Landmark design language.

## Command contract

```bash
xojo-docgen \
  -single /path/to/App.xojo_project \
  -database data/app.sqlite \
  -database /absolute/path/to/audit.sqlite \
  -out docs/api \
  -publish-prep
```

- `-database` is repeatable and requires `-single`.
- Relative paths resolve from the directory containing the `.xojo_project`.
- Files are opened through SQLite URI `mode=ro` and `PRAGMA query_only`.
- DocGen reads schema metadata only. It does not issue `SELECT` queries against
  application tables and does not emit row data.
- A non-SQLite file or unreadable table schema fails generation with the
  database path in the error.

## Extracted schema

The portable JSON contract records:

- database name, source label, file size, dialect, and `user_version`;
- tables, including `STRICT`, `WITHOUT ROWID`, and virtual-table state;
- columns, declared type, nullability, defaults, primary-key order,
  single-column uniqueness, generated/hidden state, and auto-increment state;
- indexes, ordered index columns, collation, direction, uniqueness, origin,
  partial-index state, and SQL definition;
- declared foreign keys, including composite column order and update/delete
  actions;
- views, projected columns, SQL definition, and any column-inspection error;
- triggers and their SQL definitions;
- normalized relationships used by the ER reader.

SQLite metadata comes from `sqlite_schema`, `PRAGMA table_list`,
`PRAGMA table_xinfo`, `PRAGMA index_list`, `PRAGMA index_xinfo`, and
`PRAGMA foreign_key_list`.

## Relationship truth model

The diagram never silently upgrades a guess into a constraint.

1. Foreign keys declared by SQLite are `origin: declared`. The UI renders them
   as solid lines and shows their update/delete actions.
2. A missing relationship may become `origin: suggested` only when all of the
   following are true:
   - the source column is not itself a primary-key column;
   - its normalized name matches either `<singular table><primary key>`
     (`user_id`, `CustomerID`, `ProductCode`) or the target primary-key name
     exactly (`InvoiceNo`);
   - exactly one target table satisfies that naming rule;
   - that table has exactly one primary-key column;
   - source and target SQLite type affinities are compatible;
   - the source column is not already covered by a declared foreign key.
3. Suggested relationships are dashed, say `suggested`, retain an evidence
   sentence, and are described as naming-based suggestions rather than
   constraints.
4. Ambiguous matches and incompatible types produce no edge.

This makes legacy databases readable while preserving the difference between
database truth and documentation assistance.

## Data dictionary

The dictionary is the primary detail surface:

- table and view navigation remains sticky on desktop;
- filtering matches table names, column names, types, key state, SQL, and
  indexes;
- a table route is bookmarkable, for example
  `#database/app/dictionary/invoices/`;
- columns use a dense, horizontally scrollable semantic table;
- relationships, indexes, SQL definitions, views, triggers, and view
  diagnostics remain in the same editorial hierarchy;
- light and dark modes use the generated primary-color variants;
- schema metadata is searchable together with Xojo API entities.

## ER diagram

The ER route is bookmarkable as `#database/<slug>/diagram/`. The viewer is a
documentation tool, not a schema editor.

- AntV X6 2.19.2 is pinned for compatibility with the reviewed
  `model-toolkit` implementation.
- Dagre 0.8.5 produces a deterministic directed layout.
- field-aligned ports, Manhattan routing, and rounded connectors make
  relationships traceable;
- solid edges are declared constraints and dashed edges are suggestions;
- pan, controlled zoom, fit, arrange, table focus, selection inspection, and
  dictionary drill-down are available;
- double-clicking a table opens its dictionary;
- AntV X6 and Dagre load only when the ER route is opened;
- both libraries and license files are vendored into the template, so a
  published site remains offline-capable.

For schemas up to 80 tables and 1,200 columns, nodes show their fields directly.
Larger schemas switch to compact topology nodes. All tables and edges remain on
the canvas, while the inspector and dictionary retain full column detail. This
avoids thousands of HTML field rows on the graph without hiding the topology.
Compact mode also uses the cheaper orthogonal router; smaller diagrams retain
field-to-field Manhattan routing.

## Fault tolerance

SQLite permits views that refer to missing or renamed columns. Calling
`PRAGMA table_xinfo` compiles such a view and may return an error. DocGen keeps
the view name and SQL definition, records the diagnostic as
`inspectionError`, and continues building the remaining documentation.

Table metadata errors remain fatal because an incomplete table dictionary could
misrepresent primary keys, nullability, or relationships.

## Performance reference

The feature was exercised against
`model-toolkit/imagine_hospital_rel.sqlite`:

- 252 tables;
- 6,398 columns;
- 26 views, including 6 invalid view definitions captured as diagnostics;
- 34 declared relationships;
- 371 unambiguous naming-based suggestions;
- 2.65 MB generated database JSON;
- 0.77 seconds extraction and 1.74 seconds strict MkDocs build on the development
  machine used for this feature.

The browser stress check rendered 252 compact HTML nodes and 405 edges without
creating any field-row DOM inside the canvas.

## References

- SQLite schema table: <https://www.sqlite.org/schematab.html>
- SQLite PRAGMA reference: <https://www.sqlite.org/pragma.html>
- AntV X6 ports: <https://x6.antv.antgroup.com/en/tutorial/basic/port>
- AntV X6 routers: <https://x6.antv.antgroup.com/en/api/registry/router>
- Current dependency and license notices:
  [`../THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md)
- Reviewed implementation:
  `/Users/worajedt/Xojo Projects/model-toolkit/mcp-antv-x6/examples/sakila-er-diagram.html`
