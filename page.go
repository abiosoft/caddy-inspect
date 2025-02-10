package inspect

import (
	"embed"
)

//go:embed static/assets static/index.html
var staticFS embed.FS
