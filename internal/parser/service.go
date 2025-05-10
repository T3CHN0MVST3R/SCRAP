package parser

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"website-scraper/internal/downloader"
	"website-scraper/internal/models"
	"website-scraper/internal/parser/platforms"
	"website-scraper/internal/repo"
	"website-scraper/internal/templates"
)

// parserService реализация ParserService
type parserService struct {
	repo            repo.ParserRepo
	downloader      *downloader.Downloader
	templateService *templates.TemplateService
	wordpressParser platforms.PlatformParser
	tildaParser     platforms.PlatformParser
	bitrixParser    platforms.PlatformParser
	html5Parser     *platforms.HTML5Parser // Changed to pointer type
}

// NewParserService создает новый экземпляр parserService
func NewParserService(
	repo repo.ParserRepo,
	downloader *downloader.Downloader,
	templateService *templates.TemplateService,
	wordpressParser platforms.PlatformParser,
	tildaParser platforms.PlatformParser,
	bitrixParser platforms.PlatformParser,
	html5Parser *platforms.HTML5Parser, // Changed to pointer type
) ParserService {
	return &parserService{
		repo:            repo,
		downloader:      downloader,
		templateService: templateService,
		wordpressParser: wordpressParser,
		tildaParser:     tildaParser,
		bitrixParser:    bitrixParser,
		html5Parser:     html5Parser,
	}
}

// ParseURL парсит URL и сохраняет результаты в базу данных
func (s *parserService) ParseURL(ctx context.Context, url string) (uuid.UUID, error) {
	// Создаем операцию в БД
	operationID, err := s.repo.CreateOperation(ctx, url)
	if err != nil {
		return uuid.Nil, err
	}

	// Обновляем статус операции
	err = s.repo.UpdateOperationStatus(ctx, operationID, models.StatusProcessing)
	if err != nil {
		return operationID, err
	}

	// Запускаем парсинг в отдельной горутине
	go func() {
		// Создаем новый контекст для горутины
		goCtx := context.Background()

		// Загружаем страницу
		html, err := s.downloader.DownloadPage(goCtx, url)
		if err != nil {
			s.repo.UpdateOperationStatus(goCtx, operationID, models.StatusError)
			return
		}

		// Определяем платформу сайта
		platform := s.DetectPlatform(html)

		// Парсим шапку и подвал в зависимости от платформы
		var headerBlock, footerBlock *models.Block

		switch platform {
		case models.PlatformWordPress:
			headerBlock, err = s.wordpressParser.ParseHeader(html)
			if err != nil {
				return
			}

			footerBlock, err = s.wordpressParser.ParseFooter(html)
			if err != nil {
				return
			}
		case models.PlatformTilda:
			headerBlock, err = s.tildaParser.ParseHeader(html)
			if err != nil {
				return
			}

			footerBlock, err = s.tildaParser.ParseFooter(html)
			if err != nil {
				return
			}
		case models.PlatformBitrix:
			headerBlock, err = s.bitrixParser.ParseHeader(html)
			if err != nil {
				return
			}

			footerBlock, err = s.bitrixParser.ParseFooter(html)
			if err != nil {
				return
			}
		case models.PlatformHTML5:
			var blocks []*models.Block
			templates, err := s.templateService.GetTemplates(platform)
			if err != nil {
				return
			}

			blocks, err = s.html5Parser.ParseAndClassifyPage(html, templates)
			if err != nil {
				return
			}

			// Сохраняем найденные блоки в БД
			for _, block := range blocks {
				if block != nil {
					block.OperationID = operationID

					err = s.repo.SaveBlock(goCtx, block)
					if err != nil {
						return
					}
				}
			}
		}

		// Сохраняем найденные блоки в БД
		if headerBlock != nil {
			headerBlock.OperationID = operationID
			err = s.repo.SaveBlock(goCtx, headerBlock)
			if err != nil {
				return
			}
		}

		if footerBlock != nil {
			footerBlock.OperationID = operationID
			err = s.repo.SaveBlock(goCtx, footerBlock)
			if err != nil {
				return
			}
		}

		// Обновляем статус операции
		err = s.repo.UpdateOperationStatus(goCtx, operationID, models.StatusCompleted)
		if err != nil {
			return
		}
	}()

	return operationID, nil
}

// GetOperationResult получает результаты операции по ID
func (s *parserService) GetOperationResult(ctx context.Context, operationID uuid.UUID) (*models.GetOperationResultResponse, error) {
	// Получаем операцию из БД
	operation, err := s.repo.GetOperationByID(ctx, operationID)
	if err != nil {
		return nil, err
	}

	// Получаем блоки операции
	blocks, err := s.repo.GetBlocksByOperationID(ctx, operationID)
	if err != nil {
		return nil, err
	}

	// Формируем ответ
	response := &models.GetOperationResultResponse{
		Operation: *operation,
		Blocks:    blocks,
	}

	return response, nil
}

// ExportOperation экспортирует результаты операции в файл
func (s *parserService) ExportOperation(ctx context.Context, operationID uuid.UUID, format string) ([]byte, string, error) {
	// Получаем результаты операции
	result, err := s.GetOperationResult(ctx, operationID)
	if err != nil {
		return nil, "", err
	}

	var filename string
	var content []byte

	// В зависимости от формата экспортируем результаты
	switch format {
	case "excel":
		// Создаем Excel-файл
		f := excelize.NewFile()

		// Устанавливаем заголовки для первого листа (Информация об операции)
		f.SetCellValue("Sheet1", "A1", "ID")
		f.SetCellValue("Sheet1", "B1", "URL")
		f.SetCellValue("Sheet1", "C1", "Status")
		f.SetCellValue("Sheet1", "D1", "Created At")
		f.SetCellValue("Sheet1", "E1", "Updated At")

		// Заполняем данные операции
		f.SetCellValue("Sheet1", "A2", result.Operation.ID.String())
		f.SetCellValue("Sheet1", "B2", result.Operation.URL)
		f.SetCellValue("Sheet1", "C2", result.Operation.Status)
		f.SetCellValue("Sheet1", "D2", result.Operation.CreatedAt.Format(time.RFC3339))
		f.SetCellValue("Sheet1", "E2", result.Operation.UpdatedAt.Format(time.RFC3339))

		// Создаем новый лист для блоков
		f.NewSheet("Blocks")

		// Устанавливаем заголовки для листа блоков
		f.SetCellValue("Blocks", "A1", "ID")
		f.SetCellValue("Blocks", "B1", "Type")
		f.SetCellValue("Blocks", "C1", "Platform")
		f.SetCellValue("Blocks", "D1", "Created At")
		f.SetCellValue("Blocks", "E1", "Template Name")
		f.SetCellValue("Blocks", "F1", "Components")

		// Устанавливаем ширину колонок
		f.SetColWidth("Blocks", "A", "A", 36)
		f.SetColWidth("Blocks", "B", "C", 15)
		f.SetColWidth("Blocks", "D", "D", 25)
		f.SetColWidth("Blocks", "E", "E", 30)
		f.SetColWidth("Blocks", "F", "F", 50)

		// Заполняем данные блоков
		for i, block := range result.Blocks {
			row := i + 2
			f.SetCellValue("Blocks", fmt.Sprintf("A%d", row), block.ID.String())
			f.SetCellValue("Blocks", fmt.Sprintf("B%d", row), block.BlockType)
			f.SetCellValue("Blocks", fmt.Sprintf("C%d", row), block.Platform)
			f.SetCellValue("Blocks", fmt.Sprintf("D%d", row), block.CreatedAt.Format(time.RFC3339))

			// Извлекаем название шаблона из контента, если доступно
			templateName := "N/A"
			if content, ok := block.Content.(map[string]interface{}); ok {
				if tn, ok := content["template_name"].(string); ok {
					templateName = tn
				}
			}
			f.SetCellValue("Blocks", fmt.Sprintf("E%d", row), templateName)

			// Преобразуем контент в строку JSON
			contentStr := fmt.Sprintf("%v", block.Content)
			f.SetCellValue("Blocks", fmt.Sprintf("F%d", row), contentStr)
		}

		// Сохраняем Excel-файл в буфер
		buffer, err := f.WriteToBuffer()
		if err != nil {
			return nil, "", fmt.Errorf("failed to write Excel file: %w", err)
		}

		content = buffer.Bytes()
		filename = fmt.Sprintf("operation_%s.xlsx", operationID.String())

	case "text":
		// Формируем текстовый отчет
		textContent := fmt.Sprintf("Operation ID: %s\n", result.Operation.ID.String())
		textContent += fmt.Sprintf("URL: %s\n", result.Operation.URL)
		textContent += fmt.Sprintf("Status: %s\n", result.Operation.Status)
		textContent += fmt.Sprintf("Created At: %s\n", result.Operation.CreatedAt.Format(time.RFC3339))
		textContent += fmt.Sprintf("Updated At: %s\n\n", result.Operation.UpdatedAt.Format(time.RFC3339))

		textContent += "Blocks:\n"
		for _, block := range result.Blocks {
			textContent += fmt.Sprintf("  ID: %s\n", block.ID.String())
			textContent += fmt.Sprintf("  Type: %s\n", block.BlockType)
			textContent += fmt.Sprintf("  Platform: %s\n", block.Platform)
			textContent += fmt.Sprintf("  Created At: %s\n", block.CreatedAt.Format(time.RFC3339))
			textContent += fmt.Sprintf("  HTML: %s\n\n", block.HTML)
		}

		content = []byte(textContent)
		filename = fmt.Sprintf("operation_%s.txt", operationID.String())

	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}

	return content, filename, nil
}

// DetectPlatform определяет платформу сайта по HTML
func (s *parserService) DetectPlatform(html string) models.Platform {
	if s.html5Parser.DetectPlatform(html) {
		return models.PlatformHTML5
	}

	if s.wordpressParser.DetectPlatform(html) {
		return models.PlatformWordPress
	}

	if s.tildaParser.DetectPlatform(html) {
		return models.PlatformTilda
	}

	if s.bitrixParser.DetectPlatform(html) {
		return models.PlatformBitrix
	}

	return models.PlatformUnknown
}
