// Package gowine is the module root. It exists to embed assets (migrations) that
// live at the repository root so they can be compiled into the binary.
package gowine

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS
