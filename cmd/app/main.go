package main

import (
	"go.uber.org/fx"

	"website-scraper/internal/api/routes"
	"website-scraper/internal/app"
	"website-scraper/internal/config"
	"website-scraper/internal/crawler"
	"website-scraper/internal/downloader"
	"website-scraper/internal/parser"
	"website-scraper/internal/repo"
	"website-scraper/internal/templates"
)

func main() {
	application := fx.New(
		fx.Provide(
			config.New,
		),
		repo.Module,
		templates.Module,
		parser.Module,
		downloader.Module,
		crawler.Module,
		routes.Module,
		app.Module,
	)

	application.Run()
}
