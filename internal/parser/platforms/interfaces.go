package platforms

import (
	"website-scraper/internal/models"
)

// PlatformParser представляет интерфейс для парсера конкретной платформы
type PlatformParser interface {
	// DetectPlatform проверяет, соответствует ли страница данной платформе
	DetectPlatform(html string) bool

	// ParseHeader парсит шапку сайта
	ParseHeader(html string) (*models.Block, error)

	// ParseFooter парсит подвал сайта
	ParseFooter(html string) (*models.Block, error)
}
