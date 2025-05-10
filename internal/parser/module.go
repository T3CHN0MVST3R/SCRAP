package parser

import (
	"go.uber.org/fx"

	"website-scraper/internal/downloader"
	"website-scraper/internal/parser/platforms"
	"website-scraper/internal/repo"
	"website-scraper/internal/templates"
)

// ParserDependencies groups all parser dependencies
type ParserDependencies struct {
	fx.In

	Repo            repo.ParserRepo
	Downloader      *downloader.Downloader
	TemplateService *templates.TemplateService
	WordPressParser platforms.PlatformParser `name:"wordpress"`
	TildaParser     platforms.PlatformParser `name:"tilda"`
	BitrixParser    platforms.PlatformParser `name:"bitrix"`
	HTML5Parser     *platforms.HTML5Parser
}

// Module регистрирует зависимости для парсера
var Module = fx.Module("parser",
	platforms.Module,
	fx.Provide(
		func(deps ParserDependencies) ParserService {
			return NewParserService(
				deps.Repo,
				deps.Downloader,
				deps.TemplateService,
				deps.WordPressParser,
				deps.TildaParser,
				deps.BitrixParser,
				deps.HTML5Parser,
			)
		},
	),
)
