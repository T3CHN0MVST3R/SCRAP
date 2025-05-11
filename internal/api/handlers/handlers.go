package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

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

	// Вызываем сервис для экспорта операции
	content, filename, err := h.parserService.ExportOperation(r.Context(), operationID, format)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Ошибка при экспорте операции: "+err.Error())
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", getContentType(format))
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))

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

	// Создаем экземпляр downloader
	downloader := downloader.NewDownloader(h.config)

	// Загружаем блок
	data, filename, err := downloader.DownloadBlock(operationID, blockID, format)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Блок не найден: "+err.Error())
		return
	}

	// Определяем Content-Type
	var contentType string
	if format == "html" {
		contentType = "text/html"
	} else {
		contentType = "application/json"
	}

	// Отправляем файл
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	w.WriteHeader(http.StatusOK)
	w.Write(data)
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

	// Создаем экземпляр downloader
	downloader := downloader.NewDownloader(h.config)

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

	// Определяем имя файла
	filename := filepath.Base(archivePath)

	// Устанавливаем заголовки
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// Отправляем файл
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)

	// Удаляем архив после отправки
	defer func() {
		if err := os.Remove(archivePath); err != nil {
			// Логируем ошибку, но не останавливаем выполнение
			log.Printf("Ошибка удаления архива %s: %v", archivePath, err)
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
