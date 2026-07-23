(function () {
  "use strict";

  var base = window.__XOJO_DOCGEN_BASE__ || ".";
  var assetVersion = window.__XOJO_DOCGEN_ASSET_VERSION__ || "";
  var state = { manifest: null, documents: [], databases: [], navigation: [], dark: false };
  var structuralSections = new Set([
    "Version Info", "Event Definitions", "Event Handlers", "Methods",
    "Properties", "Properties — internal", "Constants — internal", "Controls"
  ]);

  function byId(id) { return document.getElementById(id); }
  function versioned(location) {
    if (!assetVersion) return location;
    return location + (location.includes("?") ? "&" : "?") +
      "v=" + encodeURIComponent(assetVersion);
  }
  function escapeHTML(value) {
    return String(value).replace(/[&<>"']/g, function (character) {
      return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[character];
    });
  }
  function initials(name) {
    return name.split(/[\s_-]+/).filter(Boolean).slice(0, 2).map(function (part) {
      return part[0].toUpperCase();
    }).join("") || "XO";
  }
  function kindGroup(kind) {
    if (kind === "Page") return "Pages";
    if (kind === "Class") return "Classes";
    return kind.endsWith("s") ? kind : kind + "s";
  }
  function compareText(first, second) {
    return String(first || "").localeCompare(String(second || ""), undefined, {
      numeric: true,
      sensitivity: "base"
    });
  }
  function lastNamePart(name) {
    var parts = String(name || "").split(".");
    return parts[parts.length - 1] || name;
  }
  function entityLabel(entity) {
    return entity.localName || lastNamePart(entity.name);
  }
  function surfaceLabel(projectType) {
    var target = String(projectType || "").toLowerCase();
    if (target.includes("desktop")) return "Window";
    if (target.includes("web")) return "Page";
    if (target.includes("mobile") || target.includes("ios") || target.includes("android")) {
      return "Screen";
    }
    return "Surface";
  }
  function pluralize(label) {
    return label.endsWith("s") ? label : label + "s";
  }
  function navigationFor(entity) {
    if (entity.navigation && entity.navigation.category) return entity.navigation;
    if (entity.kind === "Page") return { category: "surface", section: "Surfaces" };
    if (entity.kind === "Class") return { category: "class", section: "Classes" };
    if (entity.kind === "Session") return { category: "class", section: "Sessions" };
    if (entity.kind === "Interface") return { category: "class", section: "Interfaces" };
    if (entity.kind === "Module") return { category: "library", section: "Modules" };
    if (entity.kind === "Menu Bar") return { category: "misc", section: "Menu Bars" };
    if (entity.kind === "Toolbar") return { category: "misc", section: "Toolbars" };
    return { category: "misc", section: kindGroup(entity.kind) };
  }
  function entityMenuItem(entity) {
    return {
      label: entityLabel(entity),
      detail: entity.kind + (entity.members ? " · " + entity.members : ""),
      location: entity.location,
      order: entity.kind === "Module" ? 0 : 1,
      entity: entity
    };
  }
  function ensureSection(sections, key, label, context, order) {
    var section = sections.find(function (candidate) { return candidate.key === key; });
    if (section) return section;
    section = {
      key: key,
      label: label,
      context: context || "",
      order: order || 0,
      items: []
    };
    sections.push(section);
    return section;
  }
  function buildNavigationModel(project, entities, databases) {
    var targetSurface = surfaceLabel(project.type);
    var categories = {
      surface: { id: "surface", label: targetSurface, sections: [] },
      class: { id: "class", label: "Class", sections: [] },
      database: { id: "database", label: "Database", sections: [] },
      library: { id: "library", label: "Library", sections: [] },
      misc: { id: "misc", label: "Misc", sections: [] }
    };
    var sectionOrder = {
      Surfaces: 0,
      Dialogs: 1,
      Classes: 0,
      Containers: 1,
      Sessions: 2,
      Interfaces: 3,
      "Menu Bars": 0,
      Toolbars: 1,
      Other: 2
    };

    entities.forEach(function (entity) {
      var navigation = navigationFor(entity);
      var category = categories[navigation.category] || categories.misc;
      if (category.id === "library") {
        var libraryName = entity.library || "";
        var moduleName = entity.module || (entity.kind === "Module" ? entity.name : "");
        var sectionName = lastNamePart(moduleName || libraryName || entity.name);
        var context = libraryName ? "Library · " + libraryName : "Module";
        var key = (libraryName ? "0|" + libraryName : "1") + "|" + moduleName;
        ensureSection(category.sections, key, sectionName, context, libraryName ? 0 : 1)
          .items.push(entityMenuItem(entity));
        return;
      }
      var sectionLabel = navigation.section;
      if (category.id === "surface" && sectionLabel === "Surfaces") {
        sectionLabel = pluralize(targetSurface);
      }
      ensureSection(
        category.sections,
        sectionLabel,
        sectionLabel,
        "",
        sectionOrder[navigation.section] === undefined ? 99 : sectionOrder[navigation.section]
      ).items.push(entityMenuItem(entity));
    });

    (databases || []).forEach(function (database) {
      var section = ensureSection(
        categories.database.sections,
        database.slug,
        database.name,
        database.dialect || "Database",
        0
      );
      section.items.push({
        label: "Data dictionary",
        detail: database.tables + " tables · " + database.columns + " fields",
        location: "database/" + database.slug + "/dictionary/"
      });
      section.items.push({
        label: "ER diagram",
        detail: database.relationships + " relationships",
        location: "database/" + database.slug + "/diagram/"
      });
    });

    return ["surface", "class", "database", "library", "misc"].map(function (id) {
      var category = categories[id];
      category.sections.sort(function (first, second) {
        return first.order - second.order ||
          compareText(first.context, second.context) ||
          compareText(first.label, second.label);
      });
      category.sections.forEach(function (section) {
        section.items.sort(function (first, second) {
          return (first.order || 0) - (second.order || 0) ||
            compareText(first.label, second.label);
        });
      });
      return category;
    }).filter(function (category) {
      return category.sections.some(function (section) { return section.items.length; });
    });
  }
  function renderTopNavigation(navigation) {
    var menus = navigation.map(function (category) {
      var panelId = "top-menu-" + category.id;
      var sections = category.sections.map(function (section) {
        return '<section class="top-menu-section">' +
          '<header>' + (section.context ? "<small>" + escapeHTML(section.context) + "</small>" : "") +
          "<strong>" + escapeHTML(section.label) + "</strong></header>" +
          section.items.map(function (item) {
            return '<button type="button" role="menuitem" data-location="' +
              escapeHTML(item.location) + '"><span>' + escapeHTML(item.label) +
              "</span><small>" + escapeHTML(item.detail) + "</small></button>";
          }).join("") + "</section>";
      }).join("");
      return '<div class="top-menu" data-top-menu="' + escapeHTML(category.id) + '">' +
        '<button type="button" class="top-menu-trigger" aria-expanded="false" aria-haspopup="true" ' +
        'aria-controls="' + panelId + '">' + escapeHTML(category.label) + "<i aria-hidden=\"true\"></i></button>" +
        '<div class="top-menu-panel" id="' + panelId + '" role="menu" hidden>' +
        sections + "</div></div>";
    }).join("");
    byId("top-links").innerHTML =
      '<button type="button" class="top-overview" data-home>Overview</button>' + menus;
  }
  function closeTopMenus(options) {
    document.querySelectorAll(".top-menu.is-open").forEach(function (menu) {
      menu.classList.remove("is-open");
      var trigger = menu.querySelector(".top-menu-trigger");
      var panel = menu.querySelector(".top-menu-panel");
      trigger.setAttribute("aria-expanded", "false");
      panel.hidden = true;
      if (options && options.restoreFocus) trigger.focus();
    });
  }
  function openTopMenu(menu, focusPosition) {
    closeTopMenus();
    var trigger = menu.querySelector(".top-menu-trigger");
    var panel = menu.querySelector(".top-menu-panel");
    menu.classList.add("is-open");
    trigger.setAttribute("aria-expanded", "true");
    panel.hidden = false;
    if (focusPosition) {
      var items = Array.from(panel.querySelectorAll('[role="menuitem"]'));
      var item = focusPosition === "last" ? items[items.length - 1] : items[0];
      if (item) item.focus();
    }
  }
  function bindTopNavigation() {
    document.querySelectorAll(".top-menu").forEach(function (menu) {
      var trigger = menu.querySelector(".top-menu-trigger");
      var panel = menu.querySelector(".top-menu-panel");
      trigger.addEventListener("click", function () {
        if (menu.classList.contains("is-open")) closeTopMenus();
        else openTopMenu(menu);
      });
      trigger.addEventListener("keydown", function (event) {
        if (event.key === "ArrowDown" || event.key === "ArrowUp") {
          event.preventDefault();
          openTopMenu(menu, event.key === "ArrowUp" ? "last" : "first");
        }
        if (event.key === "Escape") closeTopMenus({ restoreFocus: true });
      });
      panel.addEventListener("keydown", function (event) {
        var items = Array.from(panel.querySelectorAll('[role="menuitem"]'));
        var current = items.indexOf(document.activeElement);
        if (event.key === "Escape") {
          event.preventDefault();
          closeTopMenus({ restoreFocus: true });
          return;
        }
        if (!["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) return;
        event.preventDefault();
        var next = current;
        if (event.key === "Home") next = 0;
        if (event.key === "End") next = items.length - 1;
        if (event.key === "ArrowDown") next = (current + 1 + items.length) % items.length;
        if (event.key === "ArrowUp") next = (current - 1 + items.length) % items.length;
        if (items[next]) items[next].focus();
      });
    });
    document.addEventListener("pointerdown", function (event) {
      if (!event.target.closest(".top-menu")) closeTopMenus();
    });
  }
  function syncTopNavigation(route) {
    document.querySelectorAll(".top-overview, .top-menu-trigger").forEach(function (button) {
      button.classList.remove("active");
    });
    document.querySelectorAll(".top-menu-panel [data-location]").forEach(function (button) {
      var active = route && route.startsWith(button.dataset.location);
      button.classList.toggle("active", Boolean(active));
      if (active) {
        button.setAttribute("aria-current", "page");
        var trigger = button.closest(".top-menu").querySelector(".top-menu-trigger");
        trigger.classList.add("active");
      } else {
        button.removeAttribute("aria-current");
      }
    });
    if (!route) byId("top-links").querySelector(".top-overview").classList.add("active");
  }
  function routeLocation() {
    return decodeURIComponent(window.location.hash.replace(/^#/, ""));
  }
  function entityForRoute(route) {
    return state.manifest.entities.find(function (entity) {
      return route === entity.location || route.startsWith(entity.location);
    });
  }
  function setRoute(location) {
    var next = location ? "#" + location : window.location.pathname;
    window.history.pushState(null, "", next);
    renderRoute();
  }
  function setSection(entity, section) {
    window.history.pushState(null, "", "#" + entity.location + section);
    renderEntity(entity, section);
  }
  function documentText(documentEntry) {
    var container = document.createElement("div");
    container.innerHTML = documentEntry.text || "";
    return container.textContent || "";
  }
  function normalizeSource(html) {
    var normalized = (html || "").replace(
      /<pre><code([^>]*)>([\s\S]*?)<\/code><\/pre>/g,
      function (block, attributes, source) {
        if (!source.includes("\n")) return block;
        var lines = source.split("\n");
        var finalLine = lines.slice().reverse().find(function (line) {
          return line.trim().length > 0;
        });
        var closingMatch = finalLine && finalLine.match(
          /^(\s*)End (?:Sub|Function|Event|Class|Module)\s*$/i
        );
        var wrapperIndent = closingMatch ? closingMatch[1].length : 0;
        if (wrapperIndent === 0) return block;
        var sourceLines = lines.map(function (line, index) {
          if (index === 0 || line.trim().length === 0) return line;
          var leading = (line.match(/^ */) || [""])[0].length;
          return line.slice(Math.min(wrapperIndent, leading));
        });
        return "<pre><code" + attributes + ">" + sourceLines.join("\n") + "</code></pre>";
      }
    );
    return normalized
      .replace(/<\/pre>\s+Source\s+<pre>/g, '</pre><span class="source-label">Source</span><pre class="source-code">')
      .replace(/<pre><code(?![^>]*language-xojo)/g, '<pre><code class="language-xojo"');
  }
  function parseControlReferences(html) {
    var container = document.createElement("div");
    container.innerHTML = html || "";
    return Array.from(container.querySelectorAll("li")).map(function (item) {
      var row = item.cloneNode(true);
      row.querySelectorAll("ul, ol").forEach(function (nested) { nested.remove(); });
      var sourceLink = row.querySelector("a");
      var text = (row.textContent || "").replace(/\s+/g, " ").trim();
      var parts = text.match(/^(\S+)\s+(\S+?)(?:\s+[—-]\s+"(.*)")?$/);
      if (!parts) return null;
      return {
        typeName: parts[1],
        instanceName: parts[2],
        displayedValue: parts[3] || null,
        href: sourceLink ? sourceLink.getAttribute("href") : null
      };
    }).filter(Boolean);
  }
  function renderControlReferences(html) {
    var controls = parseControlReferences(html);
    return '<ul class="control-reference-list">' + controls.map(function (control, index) {
      var projectEntity = state.manifest.entities.find(function (entity) {
        return entity.name === control.typeName;
      });
      var typeLink;
      if (projectEntity) {
        typeLink = '<button type="button" data-control-location="' +
          escapeHTML(projectEntity.location) + '"><span>' + escapeHTML(control.typeName) +
          '</span><small>Project class</small></button>';
      } else if (control.href) {
        typeLink = '<a href="' + escapeHTML(control.href) +
          '" target="_blank" rel="noreferrer"><span>' + escapeHTML(control.typeName) +
          '</span><small>Xojo API ↗</small></a>';
      } else {
        typeLink = '<span class="control-type-name"><span>' + escapeHTML(control.typeName) +
          '</span><small>Project type</small></span>';
      }
      return '<li data-control-index="' + index + '"><span class="control-type">' + typeLink +
        '</span><strong class="control-instance">' + escapeHTML(control.instanceName) +
        '</strong><span class="control-value ' + (control.displayedValue ? "" : "is-empty") + '">' +
        (control.displayedValue ? "“" + escapeHTML(control.displayedValue) + "”" : "No initial value") +
        "</span></li>";
    }).join("") + "</ul>";
  }
  function parseVersionInfo(html) {
    var container = document.createElement("div");
    container.innerHTML = html || "";
    var month = "(?:January|February|March|April|May|June|July|August|September|October|November|December)";
    var pattern = new RegExp(
      "^(\\d+(?:\\.\\d+)*)\\s*(?:-\\s*)?(" + month +
      "\\s+(?:(?:\\d{1,2}(?:st|nd|rd|th)?,\\s*)?\\d{4}))\\s+(.+)$",
      "i"
    );
    return Array.from(container.querySelectorAll("p")).map(function (paragraph) {
      var parts = paragraph.textContent.trim().match(pattern);
      return parts ? { version: parts[1], date: parts[2], description: parts[3] } : null;
    }).filter(Boolean);
  }
  function renderVersionInfo(html) {
    var versions = parseVersionInfo(html);
    if (!versions.length) return '<div class="generated-content">' + html + "</div>";
    return '<ol class="version-info-list">' + versions.map(function (entry) {
      return '<li><div class="version-entry-meta"><strong>v' + escapeHTML(entry.version) +
        '</strong><time>' + escapeHTML(entry.date) + '</time></div><p>' +
        escapeHTML(entry.description) + "</p></li>";
    }).join("") + "</ol>";
  }
  function parseInternalMembers(html) {
    return Array.from((html || "").matchAll(
      /###\s+(.+?)\s+`([^`]+)`\s+<pre><code>([\s\S]*?)<\/code><\/pre>/g
    )).map(function (match) {
      var declaration = document.createElement("textarea");
      declaration.innerHTML = match[3];
      return {
        name: match[1].trim(),
        visibility: match[2].trim(),
        declaration: declaration.value.trim()
      };
    });
  }
  function renderInternalMembers(html) {
    var members = parseInternalMembers(html);
    if (!members.length) return '<div class="generated-content">' + html + "</div>";
    return '<div class="internal-member-list generated-content">' + members.map(function (member) {
      return '<article><header><h3>' + escapeHTML(member.name) + '</h3><span>' +
        escapeHTML(member.visibility) + '</span></header><pre><code class="language-xojo">' +
        escapeHTML(member.declaration) + "</code></pre></article>";
    }).join("") + "</div>";
  }
  function renderSectionBody(section) {
    if (section.title === "Controls") return renderControlReferences(section.text);
    if (section.title === "Version Info") return renderVersionInfo(section.text);
    if (/— internal$/.test(section.title)) return renderInternalMembers(section.text);
    return '<div class="generated-content">' + normalizeSource(section.text) + "</div>";
  }

  function renderProjectChrome() {
    var project = state.manifest.project;
    var entities = state.manifest.entities;
    var databases = state.manifest.databases || [];
    state.navigation = buildNavigationModel(project, entities, databases);
    byId("brand-mark").textContent = initials(project.name);
    byId("brand-name").textContent = project.name;
    byId("brand-target").textContent = "Xojo " + project.type + " API";
    byId("project-version").textContent = project.name + (project.version ? " " + project.version : "");
    byId("hero-kicker").textContent = project.name + " / Xojo " + project.type + " API";
    byId("hero-title").textContent = "The architecture behind a complete " + project.type.toLowerCase() + " application.";
    byId("hero-intro").textContent = "Explore the classes, sessions, pages, controls, and routines that power this Xojo project.";
    byId("featured-caption").textContent = "Generated from Xojo " + project.xojo + " · " + project.type;
    byId("footer-project").textContent = project.name + " API / Xojo " + project.xojo;
    byId("entity-count").textContent = entities.length + " documented entities";
    byId("member-count").textContent = entities.reduce(function (sum, entity) { return sum + entity.members; }, 0);
    byId("surface-count").textContent = entities.filter(function (entity) {
      return navigationFor(entity).category === "surface";
    }).length;

    var frameworkTypes = Array.from(new Set(entities.map(function (entity) {
      return entity.superName;
    }).filter(function (type) {
      return type && type !== "—";
    })));
    if (!frameworkTypes.length) frameworkTypes = ["Xojo", project.type];
    var marqueeTypes = frameworkTypes.concat(frameworkTypes);
    byId("marquee-track").innerHTML = marqueeTypes.map(function (type) {
      return "<span>" + escapeHTML(type) + "</span>";
    }).join("");
    var revealSentence = "Follow the application from startup to state, data access, interface controls, " +
      "shared logic, and the source that connects every generated entity.";
    byId("word-reveal").innerHTML = revealSentence.split(" ").map(function (word) {
      return '<span class="reveal-word">' + escapeHTML(word) + " </span>";
    }).join("");

    var facts = [
      ["Target", project.type],
      ["Xojo", project.xojo],
      ["Version", project.version],
      ["Port", project.debugPort]
    ].filter(function (fact) { return fact[1]; });
    byId("project-facts").innerHTML = facts.map(function (fact) {
      return "<div><dt>" + escapeHTML(fact[0]) + "</dt><dd>" + escapeHTML(fact[1]) + "</dd></div>";
    }).join("");

    var grouped = new Map();
    state.navigation.filter(function (category) {
      return category.id !== "database";
    }).forEach(function (category) {
      category.sections.forEach(function (section) {
        var group = section.context ?
          section.context + " / " + section.label :
          section.label;
        grouped.set(group, section.items.map(function (item) {
          return item.entity;
        }).filter(Boolean));
      });
    });
    byId("project-navigation").innerHTML = Array.from(grouped.entries()).map(function (entry) {
      return '<section class="nav-group"><h2>' + escapeHTML(entry[0]) + "</h2>" +
        entry[1].map(function (entity) {
          return '<button type="button" data-location="' + escapeHTML(entity.location) + '">' +
            "<span>" + escapeHTML(entity.name) + "</span>" +
            (entity.members ? "<small>" + entity.members + "</small>" : "") + "</button>";
        }).join("") + "</section>";
    }).join("") + (databases.length ? '<section class="nav-group nav-databases"><h2>Databases</h2>' +
      databases.map(function (database) {
        return '<button type="button" data-database-location="database/' +
          escapeHTML(database.slug) + '/dictionary/"><span>' + escapeHTML(database.name) +
          '</span><small>' + database.tables + '</small></button>';
      }).join("") + "</section>" : "");

    var representative = entities.slice(0, 12);
    byId("recent-entities").innerHTML = representative.map(function (entity) {
      return '<button type="button" data-location="' + escapeHTML(entity.location) + '">' +
        "<span>" + escapeHTML(entity.name) + "</span><small>" + escapeHTML(entity.kind) + "</small></button>";
    }).join("");

    byId("kind-accordion").innerHTML = Array.from(grouped.entries()).map(function (entry) {
      var first = entry[1][0];
      return '<button type="button" data-location="' + escapeHTML(first.location) + '">' +
        '<span class="accordion-index">' + escapeHTML(entry[0]) + "</span>" +
        '<span class="accordion-copy"><strong>' + escapeHTML(entry[0]) +
        "</strong><small>Browse generated " + escapeHTML(entry[0].toLowerCase()) + " and their source.</small></span>" +
        "<b>" + entry[1].length + "</b></button>";
    }).join("");

    renderTopNavigation(state.navigation);

    byId("action-links").innerHTML = entities.slice(0, 2).map(function (entity, index) {
      return '<button type="button" class="' + (index === 0 ? "action-light" : "action-quiet") +
        '" data-location="' + escapeHTML(entity.location) + '">Explore ' + escapeHTML(entity.name) + "</button>";
    }).join("");
    byId("open-first-entity").textContent = entities[0] ? "Open " + entities[0].name : "Open entity";

    document.querySelectorAll("[data-location]").forEach(function (button) {
      button.addEventListener("click", function () {
        closeTopMenus();
        setRoute(button.dataset.location);
      });
    });
    document.querySelectorAll("[data-home]").forEach(function (button) {
      button.addEventListener("click", function () {
        closeTopMenus();
        setRoute(null);
      });
    });
    document.querySelectorAll("[data-database-location]").forEach(function (button) {
      button.addEventListener("click", function () {
        closeTopMenus();
        setRoute(button.dataset.databaseLocation);
      });
    });
    bindTopNavigation();
  }

  function renderStories() {
    var candidates = state.documents.filter(function (documentEntry) {
      return !documentEntry.location.startsWith("database/") &&
        documentEntry.location.includes("#") &&
        /<pre><code/.test(documentEntry.text || "");
    }).slice(0, 3);
    byId("source-stories").innerHTML = candidates.map(function (entry) {
      var codeMatch = (entry.text || "").match(/<pre><code[^>]*>([\s\S]*?)<\/code><\/pre>/);
      return '<article class="story-card"><header><strong>' + escapeHTML(entry.title) +
        '</strong><span>Source</span></header><p>Generated directly from the saved Xojo project source.</p><pre><code class="language-xojo">' +
        (codeMatch ? codeMatch[1] : "") + "</code></pre></article>";
    }).join("");
    highlight(byId("source-stories"));
  }

  function renderOverview() {
    byId("overview-view").hidden = false;
    byId("entity-view").hidden = true;
    window.XojoDatabaseDocs.hide();
    document.querySelectorAll(".nav-group button").forEach(function (button) {
      button.classList.remove("active");
    });
    window.scrollTo({ top: 0 });
  }

  function renderEntity(entity, requestedSection) {
    var project = state.manifest.project;
    var overview = state.documents.find(function (entry) { return entry.location === entity.location; });
    var sections = state.documents.filter(function (entry) {
      return entry.location.startsWith(entity.location + "#");
    });
    byId("overview-view").hidden = true;
    byId("entity-view").hidden = false;
    window.XojoDatabaseDocs.hide();
    byId("entity-kind").textContent = entity.kind + " / " + entity.superName;
    byId("entity-name").textContent = entity.name;
    byId("entity-members").textContent = entity.members;
    byId("entity-target").textContent = project.type;
    byId("entity-summary").innerHTML = normalizeSource(overview ? overview.text : "");

    byId("reader-content").innerHTML = sections.map(function (section) {
      var anchor = section.location.split("#")[1] || "";
      var structural = structuralSections.has(section.title);
      var emptyGroup = structural && !(section.text || "").trim();
      return '<section class="generated-section ' + (structural ? "section-group" : "section-member") +
        (emptyGroup ? " section-group-empty" : "") +
        '" id="' + escapeHTML(anchor) + '">' +
        (structural ? "<h2>" : '<h3 class="member-heading">') + escapeHTML(section.title) +
        (structural ? "</h2>" : "</h3>") +
        renderSectionBody(section) + "</section>";
    }).join("");

    byId("reader-content").querySelectorAll("[data-control-location]").forEach(function (button) {
      button.addEventListener("click", function () { setRoute(button.dataset.controlLocation); });
    });

    byId("reader-toc").innerHTML = "<p>On this page</p>" + sections.map(function (section) {
      var anchor = section.location.split("#")[1] || "";
      return '<a class="' + (structuralSections.has(section.title) ? "toc-group" : "toc-member") +
        '" href="#' + escapeHTML(entity.location + anchor) + '" data-section="' + escapeHTML(anchor) + '">' +
        escapeHTML(section.title) + "</a>";
    }).join("");
    byId("reader-toc").querySelectorAll("[data-section]").forEach(function (link) {
      link.addEventListener("click", function (event) {
        event.preventDefault();
        setSection(entity, link.dataset.section);
        byId(link.dataset.section)?.scrollIntoView({ block: "start" });
      });
    });
    document.querySelectorAll(".nav-group button").forEach(function (button) {
      button.classList.toggle("active", button.dataset.location === entity.location);
    });
    highlight(byId("entity-view"));
    if (requestedSection) {
      requestAnimationFrame(function () {
        byId(requestedSection)?.scrollIntoView({ block: "start" });
      });
    } else {
      window.scrollTo({ top: 0 });
    }
  }

  function renderRoute() {
    var route = routeLocation();
    syncTopNavigation(route);
    if (window.XojoDatabaseDocs.renderRoute(route)) return;
    var entity = entityForRoute(route);
    if (!entity) {
      renderOverview();
      return;
    }
    var section = route.slice(entity.location.length).replace(/^#/, "");
    renderEntity(entity, section);
  }

  function highlight(container) {
    if (!container || !window.Prism) return;
    container.querySelectorAll("pre code").forEach(function (code) {
      var hasLanguage = Array.from(code.classList).some(function (className) {
        return className.startsWith("language-");
      });
      if (!hasLanguage) code.classList.add("language-xojo");
    });
    window.Prism.highlightAllUnder(container);
  }

  function renderSearch(query) {
    var needle = query.trim().toLowerCase();
    var entities = state.manifest.entities.filter(function (entity) {
      if (!needle) return true;
      var docs = state.documents.filter(function (entry) { return entry.location.startsWith(entity.location); });
      return [entity.name, entity.kind, entity.superName].join(" ").toLowerCase().includes(needle) ||
        docs.some(function (entry) {
          return (entry.title + " " + documentText(entry)).toLowerCase().includes(needle);
        });
    }).slice(0, 9);
    var databaseResults = window.XojoDatabaseDocs.search(query).slice(0, Math.max(0, 9 - entities.length));
    var resultsHTML = entities.map(function (entity) {
      return '<button type="button" data-search-location="' + escapeHTML(entity.location) + '"><span><strong>' +
        escapeHTML(entity.name) + "</strong><small>" + escapeHTML(entity.kind) + " · inherits " +
        escapeHTML(entity.superName) + "</small></span><b>Open</b></button>";
    }).join("") + databaseResults.map(function (result) {
      return '<button type="button" data-search-location="' + escapeHTML(result.location) +
        '"><span><strong>' + escapeHTML(result.title) + "</strong><small>" +
        escapeHTML(result.detail) + "</small></span><b>Open</b></button>";
    }).join("");
    byId("search-results").innerHTML = resultsHTML ||
      "<p>No API or database documentation matches “" + escapeHTML(query) + "”.</p>";
    byId("search-results").querySelectorAll("[data-search-location]").forEach(function (button) {
      button.addEventListener("click", function () {
        closeSearch();
        setRoute(button.dataset.searchLocation);
      });
    });
  }
  function openSearch() {
    byId("search-backdrop").hidden = false;
    renderSearch(byId("search-input").value);
    setTimeout(function () { byId("search-input").focus(); }, 25);
  }
  function closeSearch() { byId("search-backdrop").hidden = true; }

  function startLandmarkMotion() {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      document.querySelectorAll(".reveal-word").forEach(function (word) {
        word.style.opacity = "1";
      });
      return;
    }

    document.querySelectorAll(".hero-copy > *").forEach(function (element, index) {
      element.animate(
        [
          { transform: "translateY(38px)", opacity: 0 },
          { transform: "translateY(0)", opacity: 1 }
        ],
        { duration: 900, delay: index * 90, easing: "cubic-bezier(.22,1,.36,1)", fill: "both" }
      );
    });
    var heroVisual = document.querySelector(".hero-visual");
    if (heroVisual) {
      heroVisual.animate(
        [
          { transform: "scale(.86)", opacity: 0 },
          { transform: "scale(1)", opacity: 1 }
        ],
        { duration: 1100, delay: 200, easing: "cubic-bezier(.22,1,.36,1)", fill: "both" }
      );
    }
    var marquee = byId("marquee-track");
    if (marquee) {
      marquee.animate(
        [
          { transform: "translateX(0)" },
          { transform: "translateX(-50%)" }
        ],
        { duration: 26000, iterations: Infinity, easing: "linear" }
      );
    }

    var revealObserver = new IntersectionObserver(function (entries, observer) {
      entries.forEach(function (entry) {
        if (!entry.isIntersecting) return;
        entry.target.querySelectorAll(".reveal-word").forEach(function (word, index) {
          word.animate(
            [{ opacity: 0.16 }, { opacity: 1 }],
            { duration: 500, delay: index * 35, fill: "forwards" }
          );
        });
        observer.unobserve(entry.target);
      });
    }, { threshold: 0.25 });
    var reveal = document.querySelector(".word-reveal");
    if (reveal) revealObserver.observe(reveal);

    var storyObserver = new IntersectionObserver(function (entries, observer) {
      entries.forEach(function (entry) {
        if (!entry.isIntersecting) return;
        entry.target.animate(
          [
            { transform: "translateY(72px) scale(.9)", opacity: 0.35 },
            { transform: "translateY(0) scale(1)", opacity: 1 }
          ],
          { duration: 760, easing: "cubic-bezier(.22,1,.36,1)", fill: "both" }
        );
        observer.unobserve(entry.target);
      });
    }, { threshold: 0.2 });
    document.querySelectorAll(".story-card").forEach(function (card) {
      storyObserver.observe(card);
    });
  }

  function bindShell() {
    byId("brand-button").addEventListener("click", function () { setRoute(null); });
    byId("back-button").addEventListener("click", function () { setRoute(null); });
    byId("browse-entities").addEventListener("click", function () {
      byId("entity-bento").scrollIntoView({ behavior: "smooth" });
    });
    byId("open-first-entity").addEventListener("click", function () {
      if (state.manifest.entities[0]) setRoute(state.manifest.entities[0].location);
    });
    byId("menu-button").addEventListener("click", function () {
      if (window.matchMedia("(max-width: 900px)").matches) {
        byId("sidebar").classList.toggle("is-open");
      } else {
        byId("site-shell").classList.toggle("sidebar-collapsed");
      }
    });
    byId("search-trigger").addEventListener("click", openSearch);
    byId("search-input").addEventListener("input", function (event) { renderSearch(event.target.value); });
    byId("search-backdrop").addEventListener("mousedown", function (event) {
      if (event.target === byId("search-backdrop")) closeSearch();
    });
    byId("theme-button").addEventListener("click", function () {
      state.dark = !state.dark;
      document.documentElement.dataset.theme = state.dark ? "dark" : "light";
      byId("theme-button").innerHTML = "<span aria-hidden=\"true\">" + (state.dark ? "☀︎" : "☾") + "</span>";
      byId("theme-button").setAttribute("aria-label", state.dark ? "Use light theme" : "Use dark theme");
    });
    window.addEventListener("keydown", function (event) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        openSearch();
      }
      if (event.key === "Escape") {
        closeSearch();
        closeTopMenus();
      }
    });
    window.addEventListener("hashchange", renderRoute);
    window.addEventListener("popstate", renderRoute);
  }

  Promise.all([
    fetch(versioned(base + "/data/project.json")).then(function (response) { return response.json(); }),
    fetch(versioned(base + "/data/documents.json")).then(function (response) { return response.json(); }),
    fetch(versioned(base + "/data/databases.json")).then(function (response) { return response.json(); })
      .catch(function () { return { databases: [] }; })
  ]).then(function (results) {
    state.manifest = results[0];
    state.documents = results[1].docs || [];
    state.databases = results[2].databases || [];
    window.XojoDatabaseDocs.init({
      base: base,
      assetVersion: assetVersion,
      databases: state.databases,
      escapeHTML: escapeHTML,
      highlight: highlight,
      project: state.manifest.project,
      setRoute: setRoute
    });
    renderProjectChrome();
    renderStories();
    bindShell();
    startLandmarkMotion();
    renderRoute();
  }).catch(function (error) {
    document.body.innerHTML = "<main class=\"load-error\"><h1>Documentation could not load.</h1><pre>" +
      escapeHTML(error.message) + "</pre></main>";
  });
})();
