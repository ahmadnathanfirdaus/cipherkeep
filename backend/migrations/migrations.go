// Package migrations embeds the SQL migration files for golang-migrate.
package migrations

import "embed"

// FS holds all .sql migration files in this directory.
//
//go:embed *.sql
var FS embed.FS
