(function () {
  "use strict";

  var storageKey = "xojo-docgen-sidebar-collapsed";

  function initializeSidebarToggle() {
    var header = document.querySelector(".md-header__inner");
    var navigation = document.querySelector(".md-sidebar--primary");
    if (!header || !navigation || header.querySelector(".xojo-sidebar-toggle")) {
      return;
    }

    var button = document.createElement("button");
    button.type = "button";
    button.className = "xojo-sidebar-toggle";
    button.setAttribute("aria-controls", "xojo-primary-navigation");
    button.innerHTML = "<span></span><span></span>";

    navigation.id = "xojo-primary-navigation";
    var collapsed = window.localStorage.getItem(storageKey) === "true";

    function applyState() {
      document.body.classList.toggle("xojo-sidebar-collapsed", collapsed);
      button.setAttribute("aria-expanded", String(!collapsed));
      button.setAttribute("aria-label", collapsed ? "Show navigation sidebar" : "Hide navigation sidebar");
      button.title = collapsed ? "Show navigation sidebar" : "Hide navigation sidebar";
    }

    button.addEventListener("click", function () {
      collapsed = !collapsed;
      window.localStorage.setItem(storageKey, String(collapsed));
      applyState();
    });

    header.insertBefore(button, header.firstChild);
    applyState();
  }

  if (typeof document$ !== "undefined") {
    document$.subscribe(initializeSidebarToggle);
  } else if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initializeSidebarToggle);
  } else {
    initializeSidebarToggle();
  }
})();
