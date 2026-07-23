(function () {
  "use strict";

  var config = null;
  var graph = null;
  var graphDatabase = null;
  var registeredNode = false;
  var librariesPromise = null;
  var routePattern = /^database\/([^/]+)\/(dictionary|diagram)\/?([^/]*)\/?$/;

  function byId(id) { return document.getElementById(id); }
  function escapeHTML(value) { return config.escapeHTML(value); }
  function versioned(location) {
    if (!config.assetVersion) return location;
    return location + (location.includes("?") ? "&" : "?") +
      "v=" + encodeURIComponent(config.assetVersion);
  }
  function slugify(value) {
    return String(value).trim().toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "") || "item";
  }
  function routeFor(database, mode, table) {
    var route = "database/" + database.slug + "/" + mode + "/";
    return table ? route + slugify(table) + "/" : route;
  }
  function databaseForSlug(slug) {
    return config.databases.find(function (database) { return database.slug === slug; });
  }
  function relationshipGroups(database) {
    if (Array.isArray(database.relationships)) {
      return database.relationships.map(function (relationship) {
        var table = database.tables.find(function (entry) {
          return entry.name === relationship.fromTable;
        });
        return {
          id: relationship.id,
          table: table,
          origin: relationship.origin,
          evidence: relationship.evidence,
          onUpdate: relationship.onUpdate || "",
          onDelete: relationship.onDelete || "",
          columns: relationship.fromColumns.map(function (column, index) {
            return {
              from: column,
              targetTable: relationship.targetTable,
              targetColumn: relationship.targetColumns[index] || ""
            };
          })
        };
      }).filter(function (relationship) { return relationship.table; });
    }
    var groups = [];
    database.tables.forEach(function (table) {
      var byID = new Map();
      table.foreignKeys.forEach(function (foreignKey) {
        if (!byID.has(foreignKey.id)) byID.set(foreignKey.id, []);
        byID.get(foreignKey.id).push(foreignKey);
      });
      byID.forEach(function (columns, id) {
        groups.push({
          id: id,
          table: table,
          origin: "declared",
          evidence: "Declared by SQLite foreign-key metadata.",
          onUpdate: columns[0].onUpdate,
          onDelete: columns[0].onDelete,
          columns: columns
        });
      });
    });
    return groups;
  }
  function columnCount(database) {
    return database.tables.reduce(function (count, table) {
      return count + table.columns.length;
    }, 0);
  }
  function isLargeSchema(database) {
    return database.tables.length > 80 || columnCount(database) > 1200;
  }
  function selectedTable(database, tableSlug) {
    if (!tableSlug) return null;
    return database.tables.find(function (table) {
      return slugify(table.name) === tableSlug;
    }) || null;
  }
  function typeLabel(column) {
    return column.type || "ANY";
  }
  function columnKey(column) {
    if (column.primaryKey) return "PK" + (column.primaryKey > 1 ? " " + column.primaryKey : "");
    if (column.unique) return "UQ";
    return "";
  }
  function columnFlags(column) {
    var flags = [];
    if (!column.nullable) flags.push("required");
    if (column.generated) flags.push(column.generated + " generated");
    if (column.autoIncrement) flags.push("auto increment");
    if (column.hidden) flags.push("hidden");
    return flags.join(", ");
  }

  function hide() {
    byId("database-view").hidden = true;
    disposeGraph();
  }

  function showDatabaseShell(database, mode) {
    byId("overview-view").hidden = true;
    byId("entity-view").hidden = true;
    byId("database-view").hidden = false;
    byId("database-kind").textContent = database.dialect.toUpperCase() + " database";
    byId("database-name").textContent = database.name;
    byId("database-source").textContent = database.source;
    var stats = [
      ["Tables", database.tables.length],
      ["Columns", columnCount(database)],
      ["Relations", relationshipGroups(database).length],
      ["Views", database.views.length]
    ];
    byId("database-stats").innerHTML = stats.map(function (stat) {
      return "<div><dt>" + escapeHTML(stat[0]) + "</dt><dd>" + stat[1] + "</dd></div>";
    }).join("");
    byId("database-tabs").querySelectorAll("[data-database-tab]").forEach(function (button) {
      button.classList.toggle("active", button.dataset.databaseTab === mode);
      button.setAttribute("aria-current", button.dataset.databaseTab === mode ? "page" : "false");
    });
    document.querySelectorAll("[data-database-location]").forEach(function (button) {
      button.classList.toggle(
        "active",
        button.dataset.databaseLocation.startsWith("database/" + database.slug + "/" + mode + "/")
      );
    });
  }

  function renderRoute(route) {
    var match = route.match(routePattern);
    if (!match) return false;
    var database = databaseForSlug(match[1]);
    if (!database) return false;
    var mode = match[2];
    var table = selectedTable(database, match[3]);
    showDatabaseShell(database, mode);
    if (mode === "diagram") {
      renderDiagram(database);
    } else {
      disposeGraph();
      renderDictionary(database, table);
    }
    return true;
  }

  function renderDictionary(database, table) {
    var content = byId("database-content");
    var tableNavigation = database.tables.map(function (entry) {
      return '<button type="button" data-dictionary-table="' + escapeHTML(entry.name) +
        '" class="' + (table && table.name === entry.name ? "active" : "") + '">' +
        "<span>" + escapeHTML(entry.name) + "</span><small>" + entry.columns.length + "</small></button>";
    }).join("");
    var visibleTables = table ? [table] : database.tables;
    content.innerHTML =
      '<div class="dictionary-tools">' +
        '<label><span>Filter tables and columns</span><input id="dictionary-filter" type="search" ' +
          'placeholder="Table, column, type, or key"></label>' +
        '<p>Schema metadata only. Application row data is never read.</p>' +
      "</div>" +
      '<div class="dictionary-layout">' +
        '<aside class="dictionary-navigation"><p>Tables</p><div>' + tableNavigation + "</div>" +
          (database.views.length ? '<p class="dictionary-secondary-label">Views</p>' +
            database.views.map(function (view) {
              return '<a href="#database-view-' + escapeHTML(slugify(view.name)) + '">' +
                escapeHTML(view.name) + "</a>";
            }).join("") : "") +
        "</aside>" +
        '<div class="dictionary-content">' +
          '<div class="dictionary-summary"><strong>' + visibleTables.length +
            (visibleTables.length === 1 ? " table" : " tables") + '</strong><span>' +
            columnCountForTables(visibleTables) + " documented columns</span></div>" +
          '<div id="dictionary-tables">' + visibleTables.map(function (entry) {
            return renderTable(entry, database);
          }).join("") + "</div>" +
          (!table ? renderViews(database.views) + renderTriggers(database.triggers) : "") +
          '<div class="dictionary-empty" id="dictionary-empty" hidden>No schema items match this filter.</div>' +
        "</div>" +
      "</div>";

    content.querySelectorAll("[data-dictionary-table]").forEach(function (button) {
      button.addEventListener("click", function () {
        config.setRoute(routeFor(database, "dictionary", button.dataset.dictionaryTable));
      });
    });
    byId("dictionary-filter").addEventListener("input", function (event) {
      filterDictionary(event.target.value);
    });
    config.highlight(content);
    window.scrollTo({ top: 0 });
  }

  function columnCountForTables(tables) {
    return tables.reduce(function (count, table) { return count + table.columns.length; }, 0);
  }

  function renderTable(table, database) {
    var flags = [];
    if (table.strict) flags.push("STRICT");
    if (table.withoutRowId) flags.push("WITHOUT ROWID");
    if (table.virtual) flags.push("VIRTUAL");
    var indexes = table.indexes.map(function (index) {
      var columns = index.columns.filter(function (column) { return column.key; }).map(function (column) {
        return column.name || "expression";
      });
      return '<li><strong>' + escapeHTML(index.name) + '</strong><code>' +
        escapeHTML(columns.join(", ")) + "</code><span>" +
        [index.unique ? "unique" : "", index.partial ? "partial" : "", index.origin]
          .filter(Boolean).join(" · ") + "</span></li>";
    }).join("");
    var relationships = relationshipRows(table, database);
    return '<section class="dictionary-table" id="database-table-' + escapeHTML(slugify(table.name)) +
      '" data-dictionary-search="' + escapeHTML(tableSearchText(table)) + '">' +
      '<header><div><p>Table</p><h2>' + escapeHTML(table.name) +
      '</h2></div><div class="schema-flags">' + flags.map(function (flag) {
        return "<span>" + flag + "</span>";
      }).join("") + "</div></header>" +
      '<div class="column-table-scroll"><table class="column-table"><thead><tr>' +
        "<th>Column</th><th>Type</th><th>Key</th><th>Null</th><th>Default</th><th>Attributes</th>" +
      "</tr></thead><tbody>" + table.columns.map(function (column) {
        var key = columnKey(column);
        return '<tr data-column-search="' + escapeHTML([
          column.name, column.type, key, columnFlags(column), column.default || ""
        ].join(" ")) + '"><td><code>' + escapeHTML(column.name) + "</code></td><td><code>" +
          escapeHTML(typeLabel(column)) + '</code></td><td><span class="key-marker ' +
          (key ? "is-key" : "") + '">' + (key || "—") + "</span></td><td>" +
          (column.nullable ? "Yes" : "No") + "</td><td><code>" +
          escapeHTML(column.default === undefined ? "—" : column.default) +
          "</code></td><td>" + escapeHTML(columnFlags(column) || "—") + "</td></tr>";
      }).join("") + "</tbody></table></div>" +
      (relationships ? '<section class="schema-detail"><h3>Relationships</h3>' +
        relationships + "</section>" : "") +
      (indexes ? '<section class="schema-detail"><h3>Indexes</h3><ul class="index-list">' +
        indexes + "</ul></section>" : "") +
      (table.sql ? '<details class="schema-sql"><summary>SQL definition</summary><pre><code class="language-sql">' +
        escapeHTML(table.sql) + "</code></pre></details>" : "") +
      "</section>";
  }

  function relationshipRows(table, database) {
    var relationships = relationshipGroups(database).filter(function (relationship) {
      return relationship.table.name === table.name;
    });
    if (!relationships.length) return "";
    return '<div class="relationship-list">' + relationships.map(function (relationship) {
      var details = relationship.origin === "declared"
        ? "update " + relationship.onUpdate.toLowerCase() + " · delete " + relationship.onDelete.toLowerCase()
        : "suggested · " + relationship.evidence;
      return '<div class="' + (relationship.origin === "suggested" ? "is-suggested" : "") +
        '"><code>' + escapeHTML(relationship.columns.map(function (column) { return column.from; }).join(", ")) +
        '</code><span>references</span><strong>' + escapeHTML(relationship.columns[0].targetTable) +
        '</strong><code>' + escapeHTML(relationship.columns.map(function (column) {
          return column.targetColumn || "primary key";
        }).join(", ")) + '</code><small>' + escapeHTML(details) + "</small></div>";
    }).join("") + "</div>";
  }

  function renderViews(views) {
    if (!views.length) return "";
    return '<section class="dictionary-object-group"><header><p>Read models</p><h2>Views</h2></header>' +
      views.map(function (view) {
        return '<article id="database-view-' + escapeHTML(slugify(view.name)) +
          '" data-dictionary-search="' + escapeHTML(view.name + " " +
            view.columns.map(function (column) { return column.name + " " + column.type; }).join(" ")) +
          '"><h3>' + escapeHTML(view.name) + '</h3><p>' + view.columns.length +
          " projected columns</p>" +
          (view.inspectionError ? '<div class="schema-warning"><strong>Column metadata unavailable</strong><span>' +
            escapeHTML(view.inspectionError) + "</span></div>" : "") +
          '<pre><code class="language-sql">' + escapeHTML(view.sql || "") +
          "</code></pre></article>";
      }).join("") + "</section>";
  }

  function renderTriggers(triggers) {
    if (!triggers.length) return "";
    return '<section class="dictionary-object-group"><header><p>Database behavior</p><h2>Triggers</h2></header>' +
      triggers.map(function (trigger) {
        return '<article data-dictionary-search="' + escapeHTML(
          trigger.name + " " + trigger.table + " " + trigger.sql
        ) + '"><h3>' + escapeHTML(trigger.name) + '</h3><p>Table <code>' +
          escapeHTML(trigger.table) + '</code></p><pre><code class="language-sql">' +
          escapeHTML(trigger.sql || "") + "</code></pre></article>";
      }).join("") + "</section>";
  }

  function tableSearchText(table) {
    return [
      table.name, table.sql,
      table.columns.map(function (column) {
        return [column.name, column.type, columnKey(column), columnFlags(column)].join(" ");
      }).join(" "),
      table.indexes.map(function (index) { return index.name; }).join(" ")
    ].join(" ");
  }

  function filterDictionary(query) {
    var needle = query.trim().toLowerCase();
    var visible = 0;
    byId("dictionary-tables").querySelectorAll(".dictionary-table").forEach(function (section) {
      var showTable = !needle || section.dataset.dictionarySearch.toLowerCase().includes(needle);
      if (showTable && needle) {
        section.querySelectorAll("[data-column-search]").forEach(function (row) {
          row.hidden = !row.dataset.columnSearch.toLowerCase().includes(needle) &&
            !section.querySelector("h2").textContent.toLowerCase().includes(needle);
        });
      } else {
        section.querySelectorAll("[data-column-search]").forEach(function (row) { row.hidden = false; });
      }
      section.hidden = !showTable;
      if (showTable) visible++;
    });
    document.querySelectorAll(".dictionary-object-group article").forEach(function (article) {
      var showObject = !needle || article.dataset.dictionarySearch.toLowerCase().includes(needle);
      article.hidden = !showObject;
      if (showObject) visible++;
    });
    byId("dictionary-empty").hidden = visible > 0;
  }

  function renderDiagram(database) {
    var relations = relationshipGroups(database);
    var declaredRelations = relations.filter(function (relationship) {
      return relationship.origin === "declared";
    }).length;
    var suggestedRelations = relations.length - declaredRelations;
    byId("database-content").innerHTML =
      '<div class="er-intro"><div><p>Entity relationship diagram</p><h2>Schema topology</h2></div>' +
        '<p>Drag the canvas to pan. Use the mouse wheel with Ctrl or Command to zoom. ' +
        'Double-click a table to open its dictionary.</p></div>' +
      '<div class="er-workspace">' +
        '<div class="er-toolbar" role="toolbar" aria-label="Diagram controls">' +
          '<button type="button" id="er-zoom-out" aria-label="Zoom out">−</button>' +
          '<output id="er-zoom-value">100%</output>' +
          '<button type="button" id="er-zoom-in" aria-label="Zoom in">+</button>' +
          '<button type="button" id="er-fit">Fit schema</button>' +
          '<button type="button" id="er-layout">Arrange</button>' +
          '<label><span>Find table</span><select id="er-table-select"><option value="">Choose a table</option>' +
            database.tables.map(function (table) {
              return '<option value="' + escapeHTML(table.name) + '">' + escapeHTML(table.name) + "</option>";
            }).join("") + "</select></label>" +
        "</div>" +
        '<div class="er-stage">' +
          '<div class="er-canvas" id="er-canvas"><div class="er-loading">Preparing schema diagram…</div></div>' +
          '<aside class="er-inspector" id="er-inspector"><p>Selection</p><h3>No table selected</h3>' +
            "<span>Select a table to inspect its columns and relationships.</span></aside>" +
        "</div>" +
        '<footer class="er-status"><span>' + database.tables.length + " tables · " +
          declaredRelations + " declared · " + suggestedRelations + " suggested</span>" +
          (declaredRelations ? "<span>Solid lines are declared constraints.</span>" :
            "<strong>No declared foreign keys were found.</strong>") +
          (suggestedRelations ? "<span>Dashed lines are naming-based suggestions, not constraints.</span>" : "") +
          (isLargeSchema(database) ? "<span>Compact topology mode is active for this large schema.</span>" : "") +
        "</footer>" +
      "</div>";

    bindDiagramTools(database);
    ensureLibraries().then(function () {
      if (graphDatabase !== database.slug || !graph) buildDiagram(database);
    }).catch(function (error) {
      byId("er-canvas").innerHTML = '<div class="er-error"><strong>Diagram library could not load.</strong><span>' +
        escapeHTML(error.message) + "</span></div>";
    });
    window.scrollTo({ top: 0 });
  }

  function bindDiagramTools(database) {
    byId("er-zoom-in").addEventListener("click", function () { zoomBy(0.15); });
    byId("er-zoom-out").addEventListener("click", function () { zoomBy(-0.15); });
    byId("er-fit").addEventListener("click", fitDiagram);
    byId("er-layout").addEventListener("click", function () {
      if (!graph) return;
      arrangeDiagram(database, isLargeSchema(database));
      fitDiagram();
    });
    byId("er-table-select").addEventListener("change", function (event) {
      focusTable(database, event.target.value);
    });
  }

  function ensureLibraries() {
    if (window.X6 && window.dagre) return Promise.resolve();
    if (librariesPromise) return librariesPromise;
    librariesPromise = loadScript(versioned(config.base + "/vendor/antv-x6-2.19.2.js"))
      .then(function () { return loadScript(versioned(config.base + "/vendor/dagre-0.8.5.min.js")); })
      .then(function () {
        if (!window.X6 || !window.dagre) {
          throw new Error("AntV X6 or Dagre did not expose its browser API");
        }
      });
    return librariesPromise;
  }

  function loadScript(source) {
    return new Promise(function (resolve, reject) {
      var existing = document.querySelector('script[data-database-vendor="' + source + '"]');
      if (existing) {
        if (existing.dataset.loaded === "true") resolve();
        else existing.addEventListener("load", resolve, { once: true });
        return;
      }
      var script = document.createElement("script");
      script.src = source;
      script.dataset.databaseVendor = source;
      script.addEventListener("load", function () {
        script.dataset.loaded = "true";
        resolve();
      }, { once: true });
      script.addEventListener("error", function () {
        reject(new Error("Unable to load " + source));
      }, { once: true });
      document.head.appendChild(script);
    });
  }

  function registerDatabaseNode() {
    if (registeredNode) return;
    window.X6.Shape.HTML.register({
      shape: "xojo-database-table",
      width: 286,
      height: 84,
      html: function (cell) {
        var data = cell.getData();
        var root = document.createElement("div");
        root.className = "er-table-node" + (data.docgenCompact ? " is-summary" : "");
        root.dataset.tableName = data.name;
        root.innerHTML = '<header><span>Table</span><strong>' + escapeHTML(data.name) +
          '</strong><small>' + data.columns.length + " fields</small></header>" +
          (data.docgenCompact ? "" : "<ol>" + data.columns.map(function (column) {
            var key = columnKey(column);
            return '<li><span class="er-field-key ' + (key ? "is-key" : "") + '">' +
              (key || "·") + '</span><code>' + escapeHTML(column.name) +
              '</code><small>' + escapeHTML(typeLabel(column)) + "</small></li>";
          }).join("") + "</ol>");
        return root;
      }
    });
    registeredNode = true;
  }

  function buildDiagram(database) {
    disposeGraph();
    graphDatabase = database.slug;
    registerDatabaseNode();
    var canvas = byId("er-canvas");
    canvas.innerHTML = "";
    graph = new window.X6.Graph({
      container: canvas,
      autoResize: true,
      background: { color: "transparent" },
      grid: { visible: true, size: 16, type: "dot", args: { color: "rgba(92,109,99,.22)", thickness: 1 } },
      panning: { enabled: true },
      mousewheel: { enabled: true, modifiers: ["ctrl", "meta"], minScale: 0.24, maxScale: 1.6 },
      interacting: { nodeMovable: true, edgeMovable: false, vertexMovable: false },
      connecting: { allowBlank: false, allowLoop: false, allowNode: false }
    });

    var compact = isLargeSchema(database);
    database.tables.forEach(function (table) {
      graph.addNode({
        id: nodeID(table.name),
        shape: "xojo-database-table",
        width: 286,
        height: nodeHeight(table, compact),
        data: Object.assign({}, table, { docgenCompact: compact }),
        zIndex: 2,
        ports: tablePorts(table, compact)
      });
    });
    relationshipGroups(database).forEach(function (relationship) {
      var sourceColumn = relationship.columns[0].from;
      var targetColumn = relationship.columns[0].targetColumn ||
        primaryColumn(database, relationship.columns[0].targetTable);
      graph.addEdge({
        id: edgeID(relationship.table.name, relationship.id),
        source: {
          cell: nodeID(relationship.table.name),
          port: compact ? "summary:right" : portID(sourceColumn) + ":right"
        },
        target: {
          cell: nodeID(relationship.columns[0].targetTable),
          port: compact ? "summary:left" : portID(targetColumn)
        },
        router: compact
          ? { name: "orth" }
          : { name: "manhattan", args: { padding: 24 } },
        connector: { name: "rounded", args: { radius: 8 } },
        attrs: {
          line: {
            stroke: "var(--red)",
            strokeWidth: 1.5,
            strokeDasharray: relationship.origin === "suggested" ? "7 5" : "",
            opacity: relationship.origin === "suggested" ? 0.72 : 1,
            targetMarker: { name: "block", width: 8, height: 6 },
            sourceMarker: { name: "circle", r: 3 }
          }
        },
        labels: [{
          attrs: {
            label: {
              text: relationshipLabel(relationship) +
                (relationship.origin === "suggested" ? " · suggested" : ""),
              fill: "var(--ink-muted)",
              fontSize: 10
            },
            body: { fill: "var(--paper)", stroke: "var(--line)" }
          },
          position: 0.5
        }],
        zIndex: 1
      });
    });
    arrangeDiagram(database, compact);
    bindGraphEvents(database);
    requestAnimationFrame(fitDiagram);
  }

  function nodeHeight(table, compact) {
    return compact ? 54 : 54 + Math.max(table.columns.length, 1) * 30;
  }
  function nodeID(tableName) { return "table:" + tableName; }
  function edgeID(tableName, foreignKeyID) { return "fk:" + tableName + ":" + foreignKeyID; }
  function portID(columnName) { return "field:" + columnName; }
  function tablePorts(table, compact) {
    if (compact) {
      return {
        groups: {
          left: {
            position: "absolute",
            attrs: { circle: { r: 3, magnet: true, stroke: "var(--red)", fill: "var(--paper)" } }
          },
          right: {
            position: "absolute",
            attrs: { circle: { r: 3, magnet: true, stroke: "var(--red)", fill: "var(--paper)" } }
          }
        },
        items: [
          { id: "summary:left", group: "left", args: { x: 0, y: 27 } },
          { id: "summary:right", group: "right", args: { x: 286, y: 27 } }
        ]
      };
    }
    var items = [];
    table.columns.forEach(function (column, index) {
      var y = 54 + index * 30 + 15;
      items.push({ id: portID(column.name), group: "left", args: { x: 0, y: y } });
      items.push({ id: portID(column.name) + ":right", group: "right", args: { x: 286, y: y } });
    });
    return {
      groups: {
        left: {
          position: "absolute",
          attrs: { circle: { r: 3, magnet: true, stroke: "var(--red)", fill: "var(--paper)" } }
        },
        right: {
          position: "absolute",
          attrs: { circle: { r: 3, magnet: true, stroke: "var(--red)", fill: "var(--paper)" } }
        }
      },
      items: items
    };
  }
  function primaryColumn(database, tableName) {
    var table = database.tables.find(function (entry) { return entry.name === tableName; });
    if (!table) return "";
    var primary = table.columns.find(function (column) { return column.primaryKey === 1; });
    return primary ? primary.name : (table.columns[0] ? table.columns[0].name : "");
  }
  function relationshipLabel(relationship) {
    return relationship.columns.map(function (column) { return column.from; }).join(", ");
  }

  function arrangeDiagram(database, compact) {
    if (!graph) return;
    var layout = new window.dagre.graphlib.Graph();
    layout.setGraph({
      rankdir: database.tables.length > 14 ? "TB" : "LR",
      nodesep: 64,
      ranksep: 120,
      marginx: 40,
      marginy: 40,
      ranker: "network-simplex"
    });
    layout.setDefaultEdgeLabel(function () { return {}; });
    database.tables.forEach(function (table) {
      layout.setNode(nodeID(table.name), { width: 286, height: nodeHeight(table, compact) });
    });
    relationshipGroups(database).forEach(function (relationship) {
      layout.setEdge(
        nodeID(relationship.table.name),
        nodeID(relationship.columns[0].targetTable)
      );
    });
    window.dagre.layout(layout);
    database.tables.forEach(function (table) {
      var position = layout.node(nodeID(table.name));
      graph.getCellById(nodeID(table.name)).position(
        position.x - position.width / 2,
        position.y - position.height / 2
      );
    });
  }

  function bindGraphEvents(database) {
    graph.on("node:click", function (event) {
      showInspector(database, event.node.getData());
    });
    graph.on("node:dblclick", function (event) {
      config.setRoute(routeFor(database, "dictionary", event.node.getData().name));
    });
    graph.on("scale", updateZoomState);
  }

  function showInspector(database, table) {
    var outgoing = relationshipGroups(database).filter(function (relationship) {
      return relationship.table.name === table.name;
    });
    var incoming = relationshipGroups(database).filter(function (relationship) {
      return relationship.columns[0].targetTable === table.name;
    });
    byId("er-inspector").innerHTML =
      "<p>Selected table</p><h3>" + escapeHTML(table.name) + "</h3>" +
      '<dl><div><dt>Columns</dt><dd>' + table.columns.length + "</dd></div>" +
      "<div><dt>Outgoing</dt><dd>" + outgoing.length + "</dd></div>" +
      "<div><dt>Incoming</dt><dd>" + incoming.length + "</dd></div>" +
      "<div><dt>Indexes</dt><dd>" + table.indexes.length + "</dd></div></dl>" +
      '<ol class="er-inspector-fields">' + table.columns.slice(0, 40).map(function (column) {
        var key = columnKey(column);
        return '<li><span class="' + (key ? "is-key" : "") + '">' + (key || "·") +
          '</span><code>' + escapeHTML(column.name) + '</code><small>' +
          escapeHTML(typeLabel(column)) + "</small></li>";
      }).join("") + "</ol>" +
      (table.columns.length > 40 ? "<p class=\"er-inspector-more\">" +
        (table.columns.length - 40) + " more fields in the dictionary</p>" : "") +
      '<button type="button" id="er-open-dictionary">Open data dictionary</button>';
    byId("er-open-dictionary").addEventListener("click", function () {
      config.setRoute(routeFor(database, "dictionary", table.name));
    });
  }

  function focusTable(database, tableName) {
    if (!graph || !tableName) return;
    var node = graph.getCellById(nodeID(tableName));
    if (!node) return;
    if (graph.zoom() < 0.58 && typeof graph.zoomTo === "function") {
      graph.zoomTo(0.72);
    }
    graph.centerCell(node);
    nodeToFront(node);
    showInspector(database, node.getData());
  }
  function nodeToFront(node) {
    if (node && typeof node.toFront === "function") node.toFront();
  }
  function zoomBy(amount) {
    if (!graph) return;
    graph.zoom(amount);
    updateZoomState();
  }
  function fitDiagram() {
    if (!graph) return;
    graph.zoomToFit({ padding: 72, maxScale: 1 });
    graph.centerContent();
    updateZoomState();
  }
  function updateZoomState() {
    if (!graph || !byId("er-zoom-value")) return;
    var zoom = graph.zoom();
    byId("er-zoom-value").textContent = Math.round(zoom * 100) + "%";
    byId("er-stage")?.classList.toggle("er-lod-compact", zoom < 0.58);
  }
  function disposeGraph() {
    if (graph) {
      graph.dispose();
      graph = null;
      graphDatabase = null;
    }
  }

  function search(query) {
    var needle = query.trim().toLowerCase();
    var results = [];
    config.databases.forEach(function (database) {
      if (!needle || [database.name, database.dialect, database.source].join(" ").toLowerCase().includes(needle)) {
        results.push({
          title: database.name,
          detail: "Database · " + database.tables.length + " tables",
          location: routeFor(database, "dictionary")
        });
      }
      database.tables.forEach(function (table) {
        if (!needle || tableSearchText(table).toLowerCase().includes(needle)) {
          results.push({
            title: table.name,
            detail: database.name + " · table · " + table.columns.length + " columns",
            location: routeFor(database, "dictionary", table.name)
          });
        }
      });
    });
    return results;
  }

  function bind() {
    byId("database-back-button").addEventListener("click", function () { config.setRoute(null); });
    byId("database-tabs").querySelectorAll("[data-database-tab]").forEach(function (button) {
      button.addEventListener("click", function () {
        var route = window.location.hash.replace(/^#/, "");
        var match = route.match(routePattern);
        var database = match ? databaseForSlug(match[1]) : config.databases[0];
        if (database) config.setRoute(routeFor(database, button.dataset.databaseTab));
      });
    });
  }

  function init(options) {
    config = options;
    bind();
  }

  window.XojoDatabaseDocs = {
    hide: hide,
    init: init,
    renderRoute: renderRoute,
    routeFor: routeFor,
    search: search
  };
})();
