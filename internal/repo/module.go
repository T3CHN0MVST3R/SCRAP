package repo

import (
	"context"
	"database/sql"

	"go.uber.org/fx"

	"website-scraper/internal/config"
)

// Module регистрирует зависимости для репозиториев
var Module = fx.Module("repo",
	fx.Provide(
		func(cfg *config.Config) (*sql.DB, error) {
			db, err := sql.Open("postgres", cfg.Database.GetPostgresDSN())
			if err != nil {
				return nil, err
			}
			return db, nil
		},
		func(cfg *config.Config) (ParserRepo, error) {
			return NewPostgresRepo(cfg)
		},
		func(cfg *config.Config) (CrawlerRepo, error) {
			repo, err := NewPostgresRepo(cfg)
			return repo, err
		},
	),
	fx.Invoke(func(lc fx.Lifecycle, db *sql.DB) {
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return db.Close()
			},
		})
	}),
)
