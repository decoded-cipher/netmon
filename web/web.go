package web

import "embed"

//go:embed index.html style.css script.js
var FS embed.FS
