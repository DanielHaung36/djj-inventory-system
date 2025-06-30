// assets/embed.go
package assets

import (
	_ "embed"
)

//go:embed logo.png
var LogoPNG []byte

//go:embed invoice.tmpl
var InvoiceTmplSrc string
