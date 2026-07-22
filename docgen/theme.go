package main

import _ "embed"

// xojoGreenCSS is the Xojo green Material theme stylesheet. It is embedded at
// build time from extra.css (single source of truth) and written into each
// per-project site's stylesheets/extra.css so every site is self-contained.
//
//go:embed extra.css
var xojoGreenCSS string

// Client-side assets for source-block highlighting + fullscreen modal.
// These are vendored (not fetched at runtime) so the sites work fully offline
// and on any static host.
//
//go:embed assets/prism.js
var prismCoreJS string

//go:embed assets/xojo.prism.js
var xojoPrismJS string

//go:embed assets/source-modal.js
var sourceModalJS string
