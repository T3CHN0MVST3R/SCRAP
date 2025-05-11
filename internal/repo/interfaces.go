package repo

import (
	"context"

	"github.com/google/uuid"

	"website-scraper/internal/models"
)

// ParserRepo представляет интерфейс для репозитория парсера
type ParserRepo interface {
	// CreateOperation создает новую операцию парсинга
	CreateOperation(ctx context.Context, url string) (uuid.UUID, error)

	// UpdateOperationStatus обновляет статус операции
	UpdateOperationStatus(ctx context.Context, operationID uuid.UUID, status models.OperationStatus) error

	// GetOperationByID получает операцию по ID
	GetOperationByID(ctx context.Context, operationID uuid.UUID) (*models.Operation, error)

	// GetAllOperations получает все операции
	GetAllOperations(ctx context.Context) ([]models.Operation, error)

	// SaveBlock сохраняет блок, найденный при парсинге
	SaveBlock(ctx context.Context, block *models.Block) error

	// GetBlocksByOperationID получает все блоки по ID операции
	GetBlocksByOperationID(ctx context.Context, operationID uuid.UUID) ([]models.Block, error)

	// GetBlockByID получает блок по ID
	GetBlockByID(ctx context.Context, blockID uuid.UUID) (*models.Block, error)

	// GetAllTemplates получает все HTML теги для парсера блоков страницы
	GetAllTemplates(platform models.Platform) ([]models.BlockTemplate, error)
}

// CrawlerRepo представляет интерфейс для репозитория краулера
type CrawlerRepo interface {
	// SaveLink сохраняет ссылку
	SaveLink(ctx context.Context, link *models.Link) error

	// GetLinksByOperationID получает все ссылки по ID операции
	GetLinksByOperationID(ctx context.Context, operationID uuid.UUID) ([]models.Link, error)
}

// DBConnection интерфейс для подключения к базе данных
type DBConnection interface {
	Close() error
}
