(function () {
  "use strict";

  var base = window.__XOJO_DOCGEN_BASE__ || ".";
  var state = { manifest: null, documents: [], dark: false };
  var structuralSections = new Set([
    "Version Info", "Event Definitions", "Event Handlers", "Methods",
    "Properties", "Properties — internal", "Constants — internal", "Controls"
  ]);

  function byId(id) { return document.getElementById(id); }
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
  function routeLocation() {
    return decodeURIComponent(window.location.hash.replace(/^#/, ""));
  }
  function entityForRoute(route) {
    return state.manifest.entities.find(function (entity) {
      return route === entity.location || route.startsWith(entity.location + "#");
    });
  }
  function setRoute(location) {
    var next = location ? "#" + location : window.location.pathname;
    window.history.pushState(null, "", next);
    renderRoute();
  }
  function setSection(entity, section) {
    window.history.pushState(null, "", "#" + entity.location + "#" + section);
    renderEntity(entity, section);
  }
  function documentText(documentEntry) {
    var container = document.createElement("div");
    container.innerHTML = documentEntry.text || "";
    return container.textContent || "";
  }
  function normalizeSource(html) {
    return (html || "")
      .replace(/<\/pre>\s+Source\s+<pre>/g, '</pre><span class="source-label">Source</span><pre class="source-code">')
      .replace(/<pre><code(?![^>]*language-xojo)/g, '<pre><code class="language-xojo"');
  }

  function renderProjectChrome() {
    var project = state.manifest.project;
    var entities = state.manifest.entities;
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
      return ["Page", "Menu Bar", "Toolbar"].includes(entity.kind);
    }).length;

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
    entities.forEach(function (entity) {
      var group = kindGroup(entity.kind);
      if (!grouped.has(group)) grouped.set(group, []);
      grouped.get(group).push(entity);
    });
    byId("project-navigation").innerHTML = Array.from(grouped.entries()).map(function (entry) {
      return '<section class="nav-group"><h2>' + escapeHTML(entry[0]) + "</h2>" +
        entry[1].map(function (entity) {
          return '<button type="button" data-location="' + escapeHTML(entity.location) + '">' +
            "<span>" + escapeHTML(entity.name) + "</span>" +
            (entity.members ? "<small>" + entity.members + "</small>" : "") + "</button>";
        }).join("") + "</section>";
    }).join("");

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

    var topEntities = entities.slice(0, 2);
    byId("top-links").innerHTML = '<button type="button" data-home>Overview</button>' +
      topEntities.map(function (entity) {
        return '<button type="button" data-location="' + escapeHTML(entity.location) + '">' +
          escapeHTML(entity.name) + "</button>";
      }).join("");

    byId("action-links").innerHTML = entities.slice(0, 2).map(function (entity, index) {
      return '<button type="button" class="' + (index === 0 ? "action-light" : "action-quiet") +
        '" data-location="' + escapeHTML(entity.location) + '">Explore ' + escapeHTML(entity.name) + "</button>";
    }).join("");
    byId("open-first-entity").textContent = entities[0] ? "Open " + entities[0].name : "Open entity";

    document.querySelectorAll("[data-location]").forEach(function (button) {
      button.addEventListener("click", function () { setRoute(button.dataset.location); });
    });
    document.querySelectorAll("[data-home]").forEach(function (button) {
      button.addEventListener("click", function () { setRoute(null); });
    });
  }

  function renderStories() {
    var candidates = state.documents.filter(function (documentEntry) {
      return documentEntry.location.includes("#") && /<pre><code/.test(documentEntry.text || "");
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
    byId("entity-kind").textContent = entity.kind + " / " + entity.superName;
    byId("entity-name").textContent = entity.name;
    byId("entity-members").textContent = entity.members;
    byId("entity-target").textContent = project.type;
    byId("entity-summary").innerHTML = normalizeSource(overview ? overview.text : "");

    byId("reader-content").innerHTML = sections.map(function (section) {
      var anchor = section.location.split("#")[1] || "";
      var structural = structuralSections.has(section.title);
      return '<section class="generated-section ' + (structural ? "section-group" : "section-member") +
        '" id="' + escapeHTML(anchor) + '">' +
        (structural ? "<h2>" : '<h3 class="member-heading">') + escapeHTML(section.title) +
        (structural ? "</h2>" : "</h3>") +
        '<div class="generated-content">' + normalizeSource(section.text) + "</div></section>";
    }).join("");

    byId("reader-toc").innerHTML = "<p>On this page</p>" + sections.map(function (section) {
      var anchor = section.location.split("#")[1] || "";
      return '<a class="' + (structuralSections.has(section.title) ? "toc-group" : "toc-member") +
        '" href="#' + escapeHTML(entity.location + "#" + anchor) + '" data-section="' + escapeHTML(anchor) + '">' +
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
      code.classList.add("language-xojo");
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
    byId("search-results").innerHTML = entities.length ? entities.map(function (entity) {
      return '<button type="button" data-search-location="' + escapeHTML(entity.location) + '"><span><strong>' +
        escapeHTML(entity.name) + "</strong><small>" + escapeHTML(entity.kind) + " · inherits " +
        escapeHTML(entity.superName) + "</small></span><b>Open</b></button>";
    }).join("") : "<p>No API entities match “" + escapeHTML(query) + "”.</p>";
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

  function bindShell() {
    byId("brand-button").addEventListener("click", function () { setRoute(null); });
    byId("back-button").addEventListener("click", function () { window.history.back(); });
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
      if (event.key === "Escape") closeSearch();
    });
    window.addEventListener("hashchange", renderRoute);
    window.addEventListener("popstate", renderRoute);
  }

  Promise.all([
    fetch(base + "/data/project.json").then(function (response) { return response.json(); }),
    fetch(base + "/search/search_index.json").then(function (response) { return response.json(); })
  ]).then(function (results) {
    state.manifest = results[0];
    state.documents = results[1].docs || [];
    renderProjectChrome();
    renderStories();
    bindShell();
    renderRoute();
  }).catch(function (error) {
    document.body.innerHTML = "<main class=\"load-error\"><h1>Documentation could not load.</h1><pre>" +
      escapeHTML(error.message) + "</pre></main>";
  });
})();
