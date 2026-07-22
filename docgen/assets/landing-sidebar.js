/* Landing-page sidebar card.
   On the project landing page (index) only, this injects a card into the
   LEFT sidebar containing the project's Title, Project facts (Type, Version,
   Bundle ID, …), and Contents counts (the data that used to live in the
   center column's "## Project" and "## Contents" sections).

   The data is emitted server-side as a hidden JSON payload:
       <script type="application/json" id="lp-meta">{ … }</script>
   so this script only runs when #lp-meta is present (i.e. only on the
   landing page — never on per-entity pages). */
(function () {
  'use strict';

  function ready(fn) {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', fn);
    } else {
      fn();
    }
  }

  function escapeHTML(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function buildCard(meta) {
    var parts = [];
    parts.push('<div class="lp-meta">');

    // Title.
    if (meta.title) {
      parts.push('<div class="lp-meta__title">' + escapeHTML(meta.title) + '</div>');
    }

    // Project facts.
    if (meta.facts && meta.facts.length) {
      parts.push('<dl class="lp-meta__facts">');
      meta.facts.forEach(function (f) {
        parts.push(
          '<div class="lp-meta__fact">' +
            '<dt>' + escapeHTML(f.label) + '</dt>' +
            '<dd>' + escapeHTML(f.value) + '</dd>' +
          '</div>'
        );
      });
      parts.push('</dl>');
    }

    // Contents counts.
    if (meta.contents && meta.contents.length) {
      parts.push('<div class="lp-meta__heading">Contents</div>');
      parts.push('<table class="lp-meta__contents">');
      parts.push('<thead><tr><th>Kind</th><th>Count</th></tr></thead>');
      parts.push('<tbody>');
      meta.contents.forEach(function (k) {
        parts.push(
          '<tr><td>' + escapeHTML(k.kind) + '</td><td>' + escapeHTML(k.count) + '</td></tr>'
        );
      });
      parts.push('</tbody></table>');
    }

    parts.push('</div>');
    return parts.join('');
  }

  ready(function () {
    var payload = document.getElementById('lp-meta');
    if (!payload) return; // not the landing page

    var raw = payload.textContent || '';
    if (!raw.trim()) return;
    var meta;
    try {
      meta = JSON.parse(raw);
    } catch (e) {
      return; // malformed payload — leave the sidebar as-is
    }

    // Find the primary (left) sidebar's scrollwrap. Material names it
    // .md-sidebar--primary .md-sidebar__scrollwrap. Fall back gracefully.
    var sidebar = document.querySelector('.md-sidebar--primary .md-sidebar__scrollwrap') ||
                  document.querySelector('.md-sidebar--primary');
    if (!sidebar) return;

    var card = document.createElement('nav');
    card.className = 'lp-meta-nav';
    card.setAttribute('aria-label', 'Project overview');
    card.innerHTML = buildCard(meta);

    // Append after the site nav, so the card sits below Home / Classes / …
    sidebar.appendChild(card);
  });
})();
