package migrations

import (
	"embed"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed schema/*.sql
var migrations embed.FS

var Migrations = &migrate.EmbedFileSystemMigrationSource{
	FileSystem: migrations,
	Root:       "schema",
}
