package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"website-scraper/internal/config"
	"website-scraper/internal/models"
)

// PostgresRepo реализация интерфейсов для PostgreSQL
type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo создает новый экземпляр PostgresRepo
func NewPostgresRepo(cfg *config.Config) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", cfg.Database.GetPostgresDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepo{db: db}, nil
}

// Close закрывает соединение с базой данных
func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// CreateOperation создает новую операцию парсинга
func (r *PostgresRepo) CreateOperation(ctx context.Context, url string) (uuid.UUID, error) {
	var operationID uuid.UUID

	query := `
		INSERT INTO operations (url, status)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, url, models.StatusPending).Scan(&operationID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create operation: %w", err)
	}

	return operationID, nil
}

// UpdateOperationStatus обновляет статус операции
func (r *PostgresRepo) UpdateOperationStatus(ctx context.Context, operationID uuid.UUID, status models.OperationStatus) error {
	query := `
		UPDATE operations
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, operationID)
	if err != nil {
		return fmt.Errorf("failed to update operation status: %w", err)
	}

	return nil
}

// GetOperationByID получает операцию по ID
func (r *PostgresRepo) GetOperationByID(ctx context.Context, operationID uuid.UUID) (*models.Operation, error) {
	query := `
		SELECT id, url, status, created_at, updated_at
		FROM operations
		WHERE id = $1
	`

	var operation models.Operation
	var status string

	err := r.db.QueryRowContext(ctx, query, operationID).Scan(
		&operation.ID,
		&operation.URL,
		&status,
		&operation.CreatedAt,
		&operation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("operation not found: %s", operationID)
		}
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}

	operation.Status = models.OperationStatus(status)
	return &operation, nil
}

// SaveBlock сохраняет блок, найденный при парсинге
func (r *PostgresRepo) SaveBlock(ctx context.Context, block *models.Block) error {
	contentJSON, err := json.Marshal(block.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal block content: %w", err)
	}

	query := `
		INSERT INTO blocks (operation_id, block_type, platform, content, html)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err = r.db.QueryRowContext(
		ctx,
		query,
		block.OperationID,
		block.BlockType,
		block.Platform,
		contentJSON,
		block.HTML,
	).Scan(&block.ID, &block.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	return nil
}

// GetBlocksByOperationID получает все блоки по ID операции
func (r *PostgresRepo) GetBlocksByOperationID(ctx context.Context, operationID uuid.UUID) ([]models.Block, error) {
	query := `
		SELECT id, operation_id, block_type, platform, content, html, created_at
		FROM blocks
		WHERE operation_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks: %w", err)
	}
	defer rows.Close()

	var blocks []models.Block

	for rows.Next() {
		var block models.Block
		var blockType, platform string
		var contentJSON []byte

		err := rows.Scan(
			&block.ID,
			&block.OperationID,
			&blockType,
			&platform,
			&contentJSON,
			&block.HTML,
			&block.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}

		block.BlockType = models.BlockType(blockType)
		block.Platform = models.Platform(platform)

		if err := json.Unmarshal(contentJSON, &block.Content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal block content: %w", err)
		}

		blocks = append(blocks, block)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating blocks: %w", err)
	}

	return blocks, nil
}

// GetAllTemplates получает все HTML теги для парсера блоков страницы
func (r *PostgresRepo) GetAllTemplates(platform models.Platform) ([]models.BlockTemplate, error) {
	var rows *sql.Rows
	var err error

	switch platform {
	case models.PlatformWordPress:
		rows, err = r.db.Query("SELECT id, block_type, wordpress FROM block_templates WHERE wordpress IS NOT NULL ORDER BY (wordpress ->>'priority')::int")
	case models.PlatformTilda:
		rows, err = r.db.Query("SELECT id, block_type, tilda FROM block_templates WHERE tilda IS NOT NULL ORDER BY (tilda ->>'priority')::int")
	case models.PlatformBitrix:
		rows, err = r.db.Query("SELECT id, block_type, bitrix FROM block_templates WHERE bitrix IS NOT NULL ORDER BY (bitrix ->>'priority')::int")
	case models.PlatformHTML5:
		rows, err = r.db.Query("SELECT id, block_type, html5 FROM block_templates WHERE html5 IS NOT NULL ORDER BY (html5 ->>'priority')::int")
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.BlockTemplate

	for rows.Next() {
		var tmpl models.BlockTemplate
		var content []byte

		switch platform {
		case models.PlatformWordPress:
			err = rows.Scan(&tmpl.ID, &tmpl.BlockType, &content)
			tmpl.WordPress = json.RawMessage(content)
		case models.PlatformTilda:
			err = rows.Scan(&tmpl.ID, &tmpl.BlockType, &content)
			tmpl.Tilda = json.RawMessage(content)
		case models.PlatformBitrix:
			err = rows.Scan(&tmpl.ID, &tmpl.BlockType, &content)
			tmpl.Bitrix = json.RawMessage(content)
		case models.PlatformHTML5:
			err = rows.Scan(&tmpl.ID, &tmpl.BlockType, &content)
			tmpl.HTML5 = json.RawMessage(content)
		}

		if err != nil {
			return nil, fmt.Errorf("error scanning block template: %w", err)
		}

		templates = append(templates, tmpl)
	}

	return templates, nil
}

// SaveLink сохраняет ссылку
func (r *PostgresRepo) SaveLink(ctx context.Context, link *models.Link) error {
	query := `
		INSERT INTO links (operation_id, url, status, depth)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		link.OperationID,
		link.URL,
		link.Status,
		link.Depth,
	).Scan(&link.ID, &link.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save link: %w", err)
	}

	return nil
}

// GetLinksByOperationID получает все ссылки по ID операции
func (r *PostgresRepo) GetLinksByOperationID(ctx context.Context, operationID uuid.UUID) ([]models.Link, error) {
	query := `
		SELECT id, operation_id, url, status, depth, created_at
		FROM links
		WHERE operation_id = $1
		ORDER BY depth, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}
	defer rows.Close()

	var links []models.Link

	for rows.Next() {
		var link models.Link
		err := rows.Scan(
			&link.ID,
			&link.OperationID,
			&link.URL,
			&link.Status,
			&link.Depth,
			&link.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}

		links = append(links, link)
	}

	return links, nil
}
func (r *PostgresRepo) GetAllOperations(ctx context.Context) ([]models.Operation, error) {
	query := `
		SELECT id, url, status, created_at, updated_at
		FROM operations
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all operations: %w", err)
	}
	defer rows.Close()

	var operations []models.Operation

	for rows.Next() {
		var operation models.Operation
		var status string

		err := rows.Scan(
			&operation.ID,
			&operation.URL,
			&status,
			&operation.CreatedAt,
			&operation.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		operation.Status = models.OperationStatus(status)
		operations = append(operations, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating operations: %w", err)
	}

	return operations, nil
}

// GetBlockByID получает блок по ID
func (r *PostgresRepo) GetBlockByID(ctx context.Context, blockID uuid.UUID) (*models.Block, error) {
	query := `
		SELECT id, operation_id, block_type, platform, content, html, created_at
		FROM blocks
		WHERE id = $1
	`

	var block models.Block
	var blockType, platform string
	var contentJSON []byte

	err := r.db.QueryRowContext(ctx, query, blockID).Scan(
		&block.ID,
		&block.OperationID,
		&blockType,
		&platform,
		&contentJSON,
		&block.HTML,
		&block.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("block not found: %s", blockID)
		}
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	block.BlockType = models.BlockType(blockType)
	block.Platform = models.Platform(platform)

	if err := json.Unmarshal(contentJSON, &block.Content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block content: %w", err)
	}

	return &block, nil
}
