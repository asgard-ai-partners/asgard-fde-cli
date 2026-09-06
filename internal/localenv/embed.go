package localenv

import _ "embed"

// page is the whole UI: one file, no build step, no CDN.
//
// It is served under a Content-Security-Policy that allows it to talk to
// nothing but the process that served it, so a stylesheet or a font from
// anywhere else would silently not load. That is the right trade: this page
// holds a customer's credentials while somebody types them, and the cost of
// keeping it unable to reach the network is some hand-written CSS.
//
//go:embed page.html
var page []byte
