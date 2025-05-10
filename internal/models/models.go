package models

import (
	"time"

	"github.com/google/uuid"
)

// OperationStatus представляет статус операции парсинга
type OperationStatus string

const (
	StatusPending    OperationStatus = "pending"
	StatusProcessing OperationStatus = "processing"
	StatusCompleted  OperationStatus = "completed"
	StatusError      OperationStatus = "error"
)

// BlockType представляет тип блока
type BlockType string

const (
	BlockTypeHeader  BlockType = "header"
	BlockTypeFooter  BlockType = "footer"
	BlockTypeContent BlockType = "content"
)

// Platform представляет платформу сайта
type Platform string

const (
	PlatformWordPress Platform = "wordpress"
	PlatformTilda     Platform = "tilda"
	PlatformBitrix    Platform = "bitrix"
	PlatformHTML5     Platform = "html5"
	PlatformUnknown   Platform = "unknown"
)

// Operation представляет операцию парсинга
type Operation struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	URL       string          `json:"url" db:"url"`
	Status    OperationStatus `json:"status" db:"status"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// Block представляет блок, найденный при парсинге
type Block struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	OperationID uuid.UUID   `json:"operation_id" db:"operation_id"`
	BlockType   BlockType   `json:"block_type" db:"block_type"`
	Platform    Platform    `json:"platform" db:"platform"`
	Content     interface{} `json:"content" db:"content"`
	HTML        string      `json:"html" db:"html"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}

// BlockTemplate представляет шаблон блока
type BlockTemplate struct {
	ID        int         `json:"id" db:"id"`
	BlockType string      `json:"block_type" db:"block_type"`
	WordPress interface{} `json:"wordpress,omitempty" db:"wordpress"`
	Tilda     interface{} `json:"tilda,omitempty" db:"tilda"`
	Bitrix    interface{} `json:"bitrix,omitempty" db:"bitrix"`
	HTML5     interface{} `json:"html5,omitempty" db:"html5"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

// Link представляет ссылку, найденную краулером
type Link struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OperationID uuid.UUID `json:"operation_id" db:"operation_id"`
	URL         string    `json:"url" db:"url"`
	Status      int       `json:"status" db:"status"`
	Depth       int       `json:"depth" db:"depth"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Request/Response models
type ParseURLRequest struct {
	URL string `json:"url"`
}

type ParseURLResponse struct {
	OperationID uuid.UUID `json:"operation_id"`
}

type GetOperationResultResponse struct {
	Operation Operation `json:"operation"`
	Blocks    []Block   `json:"blocks"`
}

type ExportOperationRequest struct {
	OperationID uuid.UUID `json:"operation_id"`
	Format      string    `json:"format"` // "excel" или "text"
}

type CrawlURLRequest struct {
	URL       string `json:"url"`
	MaxDepth  int    `json:"max_depth,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
