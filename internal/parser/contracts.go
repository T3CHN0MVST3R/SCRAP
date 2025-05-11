package parser

import (
	"context"

	"github.com/google/uuid"

	"website-scraper/internal/models"
)

// ParserService представляет интерфейс для сервиса парсинга
type ParserService interface {
	// ParseURL парсит URL и сохраняет результаты в базу данных
	ParseURL(ctx context.Context, url string) (uuid.UUID, error)

	// GetOperationResult получает результаты операции по ID
	GetOperationResult(ctx context.Context, operationID uuid.UUID) (*models.GetOperationResultResponse, error)

	// ExportOperation экспортирует результаты операции в файл
	ExportOperation(ctx context.Context, operationID uuid.UUID, format string) ([]byte, string, error)

	// DetectPlatform определяет платформу сайта по HTML
	DetectPlatform(html string) models.Platform

	// GetBlocksByOperationID получает все блоки операции
	GetBlocksByOperationID(ctx context.Context, operationID uuid.UUID) ([]models.Block, error)

	// GetBlockByID получает конкретный блок по ID
	GetBlockByID(ctx context.Context, blockID uuid.UUID) (*models.Block, error)

	// SaveBlocks сохраняет блоки на диск
	SaveBlocks(ctx context.Context, blocks []models.Block) error
}
