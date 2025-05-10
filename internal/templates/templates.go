package templates

import (
	"go.uber.org/fx"

	"website-scraper/internal/models"
	"website-scraper/internal/repo"
)

// TemplateService представляет сервис для работы с шаблонами
type TemplateService struct {
	repo repo.ParserRepo
}

// NewTemplateService создает новый экземпляр TemplateService
func NewTemplateService(repo repo.ParserRepo) *TemplateService {
	return &TemplateService{
		repo: repo,
	}
}

// GetTemplates возвращает шаблоны для указанной платформы
func (s *TemplateService) GetTemplates(platform models.Platform) ([]models.BlockTemplate, error) {
	return s.repo.GetAllTemplates(platform)
}

// Module регистрирует зависимости для шаблонов
var Module = fx.Module("templates",
	fx.Provide(
		NewTemplateService,
	),
)
