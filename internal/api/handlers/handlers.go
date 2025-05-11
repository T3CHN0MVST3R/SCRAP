package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"

	"website-scraper/internal/config"
	"website-scraper/internal/crawler"
	"website-scraper/internal/downloader"
	"website-scraper/internal/models"
	"website-scraper/internal/parser"
)

// Handlers представляет набор всех обработчиков
type Handlers struct {
	config         *config.Config
	parserService  parser.ParserService
	crawlerService crawler.CrawlerService
}

// NewHandlers создает новый экземпляр Handlers
func NewHandlers(cfg *config.Config, parserService parser.ParserService, crawlerService crawler.CrawlerService) *Handlers {
	return &Handlers{
		config:         cfg,
		parserService:  parserService,
		crawlerService: crawlerService,
	}
}

// ParseURL обрабатывает запрос на парсинг URL
func (h *Handlers) ParseURL(w http.ResponseWriter, r *http.Request) {
	var req models.ParseURLRequest

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	// Проверяем URL
	if req.URL == "" {
		RespondWithError(w, http.StatusBadRequest, "URL обязателен")
		return
	}

	// Вызываем сервис для парсинга URL
	operationID, err := h.parserService.ParseURL(r.Context(), req.URL)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка при парсинге URL: "+err.Error())
		return
	}

	// Формируем ответ
	response := models.ParseURLResponse{
		OperationID: operationID,
	}

	// Добавляем информацию о доступных эндпоинтах для этой операции
	links := map[string]string{
		"get_result":  "/api/v1/operations/" + operationID.String(),
		"export":      "/api/v1/operations/" + operationID.String() + "/export",
		"download":    "/api/v1/download/" + operationID.String(),
		"save_blocks": "/api/v1/operations/" + operationID.String() + "/blocks/save",
		"blocks_list": "/api/v1/operations/" + operationID.String() + "/blocks",
	}

	// Расширенный ответ
	extendedResponse := struct {
		models.ParseURLResponse
		Links   map[string]string `json:"links"`
		Message string            `json:"message"`
	}{
		ParseURLResponse: response,
		Links:            links,
		Message:          "Операция парсинга запущена. Используйте предоставленные ссылки для получения результатов и экспорта.",
	}

	RespondWithJSON(w, http.StatusOK, extendedResponse)
}

// GetOperationResult обрабатывает запрос на получение результатов операции
func (h *Handlers) GetOperationResult(w http.ResponseWriter, r *http.Request) {
	// Получаем ID операции из URL
	vars := mux.Vars(r)
	operationIDStr := vars["id"]

	// Проверяем ID операции
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Неверный ID операции")
		return
	}

	// Вызываем сервис для получения результатов операции
	result, err := h.parserService.GetOperationResult(r.Context(), operationID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка при получении результата операции: "+err.Error())
		return
	}

	// Добавляем информацию о доступных действиях
	links := map[string]string{
		"export":      "/api/v1/operations/" + operationID.String() + "/export",
		"download":    "/api/v1/download/" + operationID.String(),
		"save_blocks": "/api/v1/operations/" + operationID.String() + "/blocks/save",
		"blocks_list": "/api/v1/operations/" + operationID.String() + "/blocks",
	}

	// Расширенный ответ
	extendedResponse := struct {
		*models.GetOperationResultResponse
		Links map[string]string `json:"links"`
	}{
		GetOperationResultResponse: result,
		Links:                      links,
	}

	RespondWithJSON(w, http.StatusOK, extendedResponse)
}

// ExportOperation обрабатывает запрос на экспорт результатов операции
func (h *Handlers) ExportOperation(w http.ResponseWriter, r *http.Request) {
	// Получаем ID операции из URL
	vars := mux.Vars(r)
	operationIDStr := vars["id"]

	// Проверяем ID операции
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Неверный ID операции")
		return
	}

	// Получаем формат экспорта из query параметров
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "excel" // По умолчанию Excel
	}

	// Проверяем формат
	if format != "excel" && format != "text" {
		RespondWithError(w, http.StatusBadRequest, "Неверный формат. Поддерживаемые форматы: excel, text")
		return
	}

	// Получаем результаты операции
	result, err := h.parserService.GetOperationResult(r.Context(), operationID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка при получении результата операции: "+err.Error())
		return
	}

	var filename string
	var content []byte

	// В зависимости от формата экспортируем результаты
	switch format {
	case "excel":
		// Создаем Excel-файл
		f := excelize.NewFile()
		defer func() {
			// Очищаем временный файл
			if err := f.Close(); err != nil {
				log.Printf("Ошибка закрытия Excel-файла: %v", err)
			}
		}()

		// Лист 1: Информация об операции
		f.SetSheetName("Sheet1", "Операция")

		// Заголовки для информации об операции
		headers := []string{"ID", "URL", "Status", "Created At", "Updated At"}
		for i, header := range headers {
			cell := string(rune('A'+i)) + "1"
			f.SetCellValue("Операция", cell, header)
		}

		// Данные операции
		f.SetCellValue("Операция", "A2", result.Operation.ID.String())
		f.SetCellValue("Операция", "B2", result.Operation.URL)
		f.SetCellValue("Операция", "C2", result.Operation.Status)
		f.SetCellValue("Операция", "D2", result.Operation.CreatedAt.Format(time.RFC3339))
		f.SetCellValue("Операция", "E2", result.Operation.UpdatedAt.Format(time.RFC3339))

		// Лист 2: Блоки
		_, err := f.NewSheet("Блоки")
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Ошибка создания листа Блоки: "+err.Error())
			return
		}

		// Заголовки для блоков
		blockHeaders := []string{"ID", "Type", "Platform", "Created At", "Content", "HTML Preview"}
		for i, header := range blockHeaders {
			cell := string(rune('A'+i)) + "1"
			f.SetCellValue("Блоки", cell, header)
		}

		// Устанавливаем ширину колонок
		f.SetColWidth("Блоки", "E", "E", 30) // Content
		f.SetColWidth("Блоки", "F", "F", 50) // HTML Preview

		// Заполняем данные блоков
		for i, block := range result.Blocks {
			row := i + 2

			// ID блока
			f.SetCellValue("Блоки", fmt.Sprintf("A%d", row), block.ID.String())

			// Тип блока
			f.SetCellValue("Блоки", fmt.Sprintf("B%d", row), block.BlockType)

			// Платформа
			f.SetCellValue("Блоки", fmt.Sprintf("C%d", row), block.Platform)

			// Дата создания
			f.SetCellValue("Блоки", fmt.Sprintf("D%d", row), block.CreatedAt.Format(time.RFC3339))

			// Контент в JSON
			contentBytes, err := json.Marshal(block.Content)
			if err != nil {
				contentBytes = []byte("{}")
			}
			f.SetCellValue("Блоки", fmt.Sprintf("E%d", row), string(contentBytes))

			// HTML (с ограничением для Excel)
			htmlContent := block.HTML
			if len(htmlContent) > 32767 { // Ограничение ячейки Excel
				htmlContent = htmlContent[:32767] + "...\n[Превышен лимит символов Excel]"
			}
			f.SetCellValue("Блоки", fmt.Sprintf("F%d", row), htmlContent)
		}

		// Лист 3: Статистика
		_, err = f.NewSheet("Статистика")
		if err != nil {
			log.Printf("Ошибка создания листа Статистика: %v", err)
		} else {
			// Статистика по типам блоков
			blockTypeStats := make(map[string]int)
			platformStats := make(map[string]int)

			for _, block := range result.Blocks {
				blockTypeStats[string(block.BlockType)]++
				platformStats[string(block.Platform)]++
			}

			// Заголовки статистики
			f.SetCellValue("Статистика", "A1", "Статистика операции")
			f.SetCellValue("Статистика", "A3", "Всего блоков:")
			f.SetCellValue("Статистика", "B3", len(result.Blocks))

			// Статистика по типам
			row := 5
			f.SetCellValue("Статистика", "A5", "По типам блоков:")
			for blockType, count := range blockTypeStats {
				row++
				f.SetCellValue("Статистика", fmt.Sprintf("A%d", row), blockType)
				f.SetCellValue("Статистика", fmt.Sprintf("B%d", row), count)
			}

			// Статистика по платформам
			row += 2
			f.SetCellValue("Статистика", fmt.Sprintf("A%d", row), "По платформам:")
			for platform, count := range platformStats {
				row++
				f.SetCellValue("Статистика", fmt.Sprintf("A%d", row), platform)
				f.SetCellValue("Статистика", fmt.Sprintf("B%d", row), count)
			}
		}

		// Установить активный лист
		f.SetActiveSheet(0)

		// Сохраняем Excel-файл в буфер
		buffer, err := f.WriteToBuffer()
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Ошибка записи Excel-файла: "+err.Error())
			return
		}

		content = buffer.Bytes()
		filename = fmt.Sprintf("operation_report_%s.xlsx", operationID.String())

	case "text":
		// Формируем полный текстовый отчет
		var textBuilder strings.Builder

		textBuilder.WriteString("=" + strings.Repeat("=", 50) + "=\n")
		textBuilder.WriteString("ОТЧЕТ ПО ОПЕРАЦИИ ПАРСИНГА\n")
		textBuilder.WriteString("=" + strings.Repeat("=", 50) + "=\n\n")

		// Информация об операции
		textBuilder.WriteString("ОСНОВНАЯ ИНФОРМАЦИЯ:\n")
		textBuilder.WriteString(fmt.Sprintf("  ID операции: %s\n", result.Operation.ID.String()))
		textBuilder.WriteString(fmt.Sprintf("  URL: %s\n", result.Operation.URL))
		textBuilder.WriteString(fmt.Sprintf("  Статус: %s\n", result.Operation.Status))
		textBuilder.WriteString(fmt.Sprintf("  Создан: %s\n", result.Operation.CreatedAt.Format("2006-01-02 15:04:05")))
		textBuilder.WriteString(fmt.Sprintf("  Обновлен: %s\n\n", result.Operation.UpdatedAt.Format("2006-01-02 15:04:05")))

		// Статистика
		textBuilder.WriteString("СТАТИСТИКА:\n")
		textBuilder.WriteString(fmt.Sprintf("  Всего найдено блоков: %d\n\n", len(result.Blocks)))

		// Статистика по типам
		blockTypeStats := make(map[string]int)
		platformStats := make(map[string]int)

		for _, block := range result.Blocks {
			blockTypeStats[string(block.BlockType)]++
			platformStats[string(block.Platform)]++
		}

		textBuilder.WriteString("  По типам блоков:\n")
		for blockType, count := range blockTypeStats {
			textBuilder.WriteString(fmt.Sprintf("    - %-10s: %d\n", blockType, count))
		}

		textBuilder.WriteString("\n  По платформам:\n")
		for platform, count := range platformStats {
			textBuilder.WriteString(fmt.Sprintf("    - %-10s: %d\n", platform, count))
		}

		textBuilder.WriteString("\n" + strings.Repeat("-", 60) + "\n\n")

		// Детальная информация о блоках
		textBuilder.WriteString("ДЕТАЛЬНАЯ ИНФОРМАЦИЯ О БЛОКАХ:\n\n")

		for i, block := range result.Blocks {
			textBuilder.WriteString(fmt.Sprintf("БЛОК #%d\n", i+1))
			textBuilder.WriteString(strings.Repeat("-", 30) + "\n")
			textBuilder.WriteString(fmt.Sprintf("ID: %s\n", block.ID.String()))
			textBuilder.WriteString(fmt.Sprintf("Тип: %s\n", block.BlockType))
			textBuilder.WriteString(fmt.Sprintf("Платформа: %s\n", block.Platform))
			textBuilder.WriteString(fmt.Sprintf("Создан: %s\n", block.CreatedAt.Format("2006-01-02 15:04:05")))

			// Контент в JSON
			textBuilder.WriteString("\nКОНТЕНТ (JSON):\n")
			contentJSON, err := json.MarshalIndent(block.Content, "  ", "  ")
			if err != nil {
				textBuilder.WriteString("  Ошибка сериализации контента\n")
			} else {
				if len(contentJSON) > 500 {
					textBuilder.WriteString(string(contentJSON[:500]))
					textBuilder.WriteString("...\n  [Показаны первые 500 символов]\n")
				} else {
					textBuilder.WriteString(string(contentJSON) + "\n")
				}
			}

			// HTML контент
			textBuilder.WriteString("\nHTML КОНТЕНТ:\n")
			if len(block.HTML) == 0 {
				textBuilder.WriteString("  [Пусто]\n")
			} else if len(block.HTML) > 1000 {
				textBuilder.WriteString(block.HTML[:1000])
				textBuilder.WriteString("...\n  [Показаны первые 1000 символов]\n")
			} else {
				textBuilder.WriteString(block.HTML + "\n")
			}

			textBuilder.WriteString("\n" + strings.Repeat("=", 60) + "\n\n")
		}

		// Подвал отчета
		textBuilder.WriteString("Отчет сгенерирован: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
		textBuilder.WriteString("Сервис: Website Scraper\n")

		content = []byte(textBuilder.String())
		filename = fmt.Sprintf("operation_report_%s.txt", operationID.String())

	default:
		RespondWithError(w, http.StatusBadRequest, "Неподдерживаемый формат")
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", getContentType(format))
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))

	// Отправляем файл
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// CrawlURL обрабатывает запрос на обход URL и сбор ссылок
func (h *Handlers) CrawlURL(w http.ResponseWriter, r *http.Request) {
	var req models.CrawlURLRequest

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	// Проверяем URL
	if req.URL == "" {
		RespondWithError(w, http.StatusBadRequest, "URL обязателен")
		return
	}

	// Устанавливаем глубину обхода
	maxDepth := h.config.Scraper.MaxDepth // По умолчанию из конфига
	if req.MaxDepth > 0 {
		maxDepth = req.MaxDepth
	}

	// Устанавливаем пользовательский User-Agent, если указан
	if req.UserAgent != "" {
		h.crawlerService.SetUserAgent(req.UserAgent)
	}

	// Устанавливаем максимальную глубину обхода
	h.crawlerService.SetMaxDepth(maxDepth)

	// Проверяем, разрешен ли домен
	if !h.crawlerService.IsAllowedDomain(req.URL) {
		RespondWithError(w, http.StatusBadRequest, "Домен не разрешен для обхода")
		return
	}

	// Получаем параметры настройки из query параметров
	concurrency := h.config.Scraper.Concurrency // По умолчанию из конфига
	if concurrencyStr := r.URL.Query().Get("concurrency"); concurrencyStr != "" {
		if c, err := strconv.Atoi(concurrencyStr); err == nil && c > 0 {
			concurrency = c
		}
	}

	// Устанавливаем дополнительные настройки краулера, если есть функции
	if crawler, ok := h.crawlerService.(interface {
		SetConcurrency(int)
	}); ok {
		crawler.SetConcurrency(concurrency)
	}

	// Обходим URL
	links, err := h.crawlerService.CrawlURL(r.Context(), req.URL, maxDepth)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка при обходе URL: "+err.Error())
		return
	}

	// Формируем ответ
	response := struct {
		URL   string   `json:"url"`
		Links []string `json:"links"`
		Count int      `json:"count"`
	}{
		URL:   req.URL,
		Links: links,
		Count: len(links),
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// DownloadByID обрабатывает запрос на загрузку файлов по ID операции
func (h *Handlers) DownloadByID(w http.ResponseWriter, r *http.Request) {
	// Получаем ID операции из URL
	vars := mux.Vars(r)
	operationIDStr := vars["id"]

	// Проверяем ID операции
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Неверный ID операции")
		return
	}

	// Получаем формат из query параметров
	format := r.URL.Query().Get("format")
	if format == "" {
		// По умолчанию Excel
		format = "excel"
	}

	// Проверяем поддерживаемый формат
	formats := []string{"excel", "text"}
	formatSupported := false
	for _, f := range formats {
		if f == format {
			formatSupported = true
			break
		}
	}

	if !formatSupported {
		RespondWithError(w, http.StatusBadRequest,
			"Неподдерживаемый формат. Поддерживаемые форматы: "+
				getFormatsList(formats))
		return
	}

	var filename string
	var contentType string

	switch format {
	case "excel":
		filename = "operation_" + operationID.String() + ".xlsx"
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "text":
		filename = "operation_" + operationID.String() + ".txt"
		contentType = "text/plain"
	}

	// Используем экспорт для получения содержимого файла
	content, _, err := h.parserService.ExportOperation(r.Context(), operationID, format)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка при загрузке файлов: "+err.Error())
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))

	// Отправляем файл клиенту
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// GetFormats обрабатывает запрос на получение доступных форматов
func (h *Handlers) GetFormats(w http.ResponseWriter, r *http.Request) {
	formats := []string{"excel", "text"}

	response := struct {
		Formats []string `json:"formats"`
	}{
		Formats: formats,
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// DownloadBlock загружает конкретный блок
func (h *Handlers) DownloadBlock(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из URL
	vars := mux.Vars(r)
	operationIDStr := vars["operation_id"]
	blockIDStr := vars["block_id"]

	// Парсим UUID
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Неверный ID операции")
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Неверный ID блока")
		return
	}

	// Получаем формат из query параметров
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "html"
	}

	// Проверяем формат
	if format != "html" && format != "json" {
		RespondWithError(w, http.StatusBadRequest, "Неподдерживаемый формат. Используйте: html, json")
		return
	}

	// Получаем блок из базы данных
	block, err := h.parserService.GetBlockByID(r.Context(), blockID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Блок не найден: "+err.Error())
		return
	}

	// Проверяем, что блок принадлежит указанной операции
	if block.OperationID != operationID {
		RespondWithError(w, http.StatusNotFound, "Блок не найден в указанной операции")
		return
	}

	var data []byte
	var filename string
	var contentType string

	if format == "html" {
		// Создаем полный HTML документ
		htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Блок %s - %s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; padding: 20px; }
        .metadata { background: #f5f5f5; padding: 15px; margin-bottom: 20px; border-radius: 5px; }
        .content { border: 1px solid #ddd; padding: 15px; margin-top: 20px; }
        pre { background: #f8f8f8; padding: 10px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="metadata">
        <h2>Информация о блоке</h2>
        <ul>
            <li><strong>ID блока:</strong> %s</li>
            <li><strong>ID операции:</strong> %s</li>
            <li><strong>Тип блока:</strong> %s</li>
            <li><strong>Платформа:</strong> %s</li>
            <li><strong>Создан:</strong> %s</li>
        </ul>
        <h3>Контент (JSON):</h3>
        <pre>%s</pre>
    </div>
    <div class="content">
        <h2>HTML содержимое блока</h2>
        %s
    </div>
</body>
</html>`,
			block.BlockType, block.ID.String(),
			block.ID.String(),
			block.OperationID.String(),
			block.BlockType,
			block.Platform,
			block.CreatedAt.Format("2006-01-02 15:04:05"),
			jsonPretty(block.Content),
			block.HTML)

		data = []byte(htmlContent)
		filename = fmt.Sprintf("block_%s_%s.html", block.BlockType, block.ID.String())
		contentType = "text/html"
	} else {
		// JSON формат
		jsonData, err := json.MarshalIndent(map[string]interface{}{
			"id":           block.ID,
			"operation_id": block.OperationID,
			"block_type":   block.BlockType,
			"platform":     block.Platform,
			"created_at":   block.CreatedAt,
			"content":      block.Content,
			"html":         block.HTML,
		}, "", "  ")
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Ошибка сериализации данных")
			return
		}

		data = jsonData
		filename = fmt.Sprintf("block_%s_%s.json", block.BlockType, block.ID.String())
		contentType = "application/json"
	}

	// Отправляем файл
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Вспомогательная функция для красивого форматирования JSON
func jsonPretty(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// GetBlockFiles возвращает список файлов блоков для операции
func (h *Handlers) GetBlockFiles(w http.ResponseWriter, r *http.Request) {
	// Получаем ID операции из URL
	vars := mux.Vars(r)
	operationIDStr := vars["operation_id"]

	// Парсим UUID
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Неверный ID операции")
		return
	}

	// Создаем экземпляр downloader
	downloader := downloader.NewDownloader(h.config)

	// Получаем список файлов
	files, err := downloader.GetBlockFiles(operationID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Файлы блоков не найдены: "+err.Error())
		return
	}

	// Формируем ответ
	response := struct {
		OperationID string              `json:"operation_id"`
		Files       map[string][]string `json:"files"`
		Total       int                 `json:"total"`
	}{
		OperationID: operationID.String(),
		Files:       files,
		Total:       len(files["html"]) + len(files["metadata"]),
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// DownloadAllBlocks создает и отправляет архив со всеми блоками операции
func (h *Handlers) DownloadAllBlocks(w http.ResponseWriter, r *http.Request) {
	// Получаем ID операции из URL
	vars := mux.Vars(r)
	operationIDStr := vars["operation_id"]

	// Парсим UUID
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Неверный ID операции")
		return
	}

	// Сначала получаем все блоки для этой операции
	blocks, err := h.parserService.GetBlocksByOperationID(r.Context(), operationID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка получения блоков: "+err.Error())
		return
	}

	// Создаем экземпляр downloader
	downloader := downloader.NewDownloader(h.config)

	// Убеждаемся, что блоки сохранены на диске
	if err := downloader.SaveBlocks(blocks); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка сохранения блоков: "+err.Error())
		return
	}

	// Создаем сводный файл, который перечисляет все шапки и подвалы по платформам
	summaryPath, err := downloader.CreateBlocksSummary(operationID, blocks)
	if err != nil {
		log.Printf("Предупреждение: Не удалось создать сводку блоков: %v", err)
		// Продолжаем, даже если создание сводки не удалось
	}

	// Создаем архив
	archivePath, err := downloader.ZipBlocks(operationID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка создания архива: "+err.Error())
		return
	}

	// Открываем архив
	file, err := os.Open(archivePath)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка открытия архива: "+err.Error())
		return
	}
	defer file.Close()

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка получения информации о файле: "+err.Error())
		return
	}

	// Устанавливаем заголовки
	filename := filepath.Base(archivePath)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// Отправляем файл
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)

	// Очистка
	defer func() {
		if err := os.Remove(archivePath); err != nil {
			log.Printf("Ошибка удаления архива %s: %v", archivePath, err)
		}
		if summaryPath != "" {
			if err := os.Remove(summaryPath); err != nil {
				log.Printf("Ошибка удаления файла сводки %s: %v", summaryPath, err)
			}
		}
	}()
}

// SaveBlocksEndpoint сохраняет блоки операции в файлы
func (h *Handlers) SaveBlocksEndpoint(w http.ResponseWriter, r *http.Request) {
	// Получаем ID операции из URL
	vars := mux.Vars(r)
	operationIDStr := vars["operation_id"]

	// Парсим UUID
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Неверный ID операции")
		return
	}

	// Получаем блоки операции
	blocks, err := h.parserService.GetBlocksByOperationID(r.Context(), operationID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка получения блоков: "+err.Error())
		return
	}

	// Создаем экземпляр downloader
	downloader := downloader.NewDownloader(h.config)

	// Сохраняем блоки
	if err := downloader.SaveBlocks(blocks); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка сохранения блоков: "+err.Error())
		return
	}

	// Формируем ответ
	response := struct {
		OperationID string            `json:"operation_id"`
		BlocksCount int               `json:"blocks_count"`
		Message     string            `json:"message"`
		Directory   string            `json:"directory"`
		Links       map[string]string `json:"links"`
	}{
		OperationID: operationID.String(),
		BlocksCount: len(blocks),
		Message:     "Блоки успешно сохранены",
		Directory:   filepath.Join(h.config.Downloader.OutputDir, "blocks", operationID.String()),
		Links: map[string]string{
			"list_files":     "/api/v1/operations/" + operationID.String() + "/blocks",
			"download_all":   "/api/v1/operations/" + operationID.String() + "/blocks/download",
			"operation_info": "/api/v1/operations/" + operationID.String(),
		},
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// getContentType возвращает Content-Type в зависимости от формата
func getContentType(format string) string {
	switch format {
	case "excel":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "text":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// getFormatsList возвращает строку с перечислением форматов
func getFormatsList(formats []string) string {
	if len(formats) == 0 {
		return ""
	}

	result := formats[0]
	for i := 1; i < len(formats); i++ {
		result += ", " + formats[i]
	}

	return result
}
