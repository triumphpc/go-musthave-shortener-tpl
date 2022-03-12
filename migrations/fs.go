// Package migrations implement run migration FS files
package migrations

import "embed"

//go:embed *.sql
var EmbedMigrations embed.FS
