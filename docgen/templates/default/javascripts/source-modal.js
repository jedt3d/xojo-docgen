/* Source block fullscreen modal.
   Adds an "expand" button to each <details class="source"> summary that opens
   the code in a near-fullscreen overlay with 20px padding.

   No custom close button — Material's copy-code button stays functional in the
   modal. Close via: click outside the code, or ESC. */
(function () {
  'use strict';

  function openModal(pre) {
    var overlay = document.createElement('div');
    overlay.className = 'source-modal-overlay';

    var wrap = document.createElement('div');
    wrap.className = 'source-modal';

    // Clone the <pre> verbatim, INCLUDING Material's injected copy-code
    // button (nav.md-code__nav) so the modal keeps that feature intact.
    var copy = pre.cloneNode(true);
    copy.className = pre.className;
    wrap.appendChild(copy);
    overlay.appendChild(wrap);
    document.body.appendChild(overlay);
    document.body.style.overflow = 'hidden';

    function close() {
      overlay.remove();
      document.body.style.overflow = '';
      document.removeEventListener('keydown', onKey);
    }
    function onKey(e) {
      if (e.key === 'Escape') close();
    }
    // Click outside the code (on the overlay itself) closes the modal.
    overlay.addEventListener('click', function (e) {
      if (e.target === overlay) close();
    });
    // Prevent clicks inside the code/nav from bubbling up and closing.
    wrap.addEventListener('click', function (e) {
      e.stopPropagation();
    });
    document.addEventListener('keydown', onKey);
  }

  function init() {
    var blocks = document.querySelectorAll('details.source');
    blocks.forEach(function (details) {
      var summary = details.querySelector('summary');
      if (!summary || summary.querySelector('.source-expand')) return;

      var btn = document.createElement('button');
      btn.className = 'source-expand';
      btn.setAttribute('aria-label', 'Expand source to fullscreen');
      btn.setAttribute('title', 'Expand to fullscreen');
      btn.innerHTML = '&#10509;'; // ⤢
      btn.addEventListener('click', function (e) {
        e.preventDefault();
        e.stopPropagation();
        var pre = details.querySelector('pre');
        if (pre) openModal(pre);
      });
      summary.appendChild(btn);
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
