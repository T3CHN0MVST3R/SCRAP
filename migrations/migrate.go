package migrations

import (
	"database/sql"
	"embed"

	"github.com/pkg/errors"
	"github.com/vnlozan/goose/v3"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return errors.Wrap(err, "failed to set dialect")
	}

	if err := goose.Up(db, "sql"); err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	return nil
}
