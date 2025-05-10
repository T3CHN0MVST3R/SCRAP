package platforms

import (
	"go.uber.org/fx"
)

// Module регистрирует зависимости для парсеров платформ
var Module = fx.Module("platforms",
	fx.Provide(
		fx.Annotate(
			NewWordPressParser,
			fx.ResultTags(`name:"wordpress"`),
		),
		fx.Annotate(
			NewTildaParser,
			fx.ResultTags(`name:"tilda"`),
		),
		fx.Annotate(
			NewBitrixParser,
			fx.ResultTags(`name:"bitrix"`),
		),
		NewHTML5Parser,
	),
)
