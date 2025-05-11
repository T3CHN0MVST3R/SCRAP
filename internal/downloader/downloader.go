package downloader

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/fx"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"

	"website-scraper/internal/config"
	"website-scraper/internal/models"
)

// Downloader представляет сервис для загрузки веб-страниц и блоков
type Downloader struct {
	cfg *config.Config
}

// NewDownloader создает новый экземпляр Downloader
func NewDownloader(cfg *config.Config) *Downloader {
	// Создаем директории для сохранения блоков
	directories := []string{
		cfg.Downloader.OutputDir,
		filepath.Join(cfg.Downloader.OutputDir, "blocks"),
		filepath.Join(cfg.Downloader.OutputDir, "html"),
		filepath.Join(cfg.Downloader.OutputDir, "exports"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Ошибка создания директории %s: %v", dir, err)
		}
	}

	return &Downloader{
		cfg: cfg,
	}
}

// DownloadPage загружает страницу и возвращает HTML
func (d *Downloader) DownloadPage(ctx context.Context, url string) (string, error) {
	// Определяем путь к браузеру
	var execPath string

	// Проверяем, какой браузер установлен
	if _, err := os.Stat("/usr/bin/google-chrome"); err == nil {
		execPath = "/usr/bin/google-chrome"
	} else if _, err := os.Stat("/usr/bin/chromium-browser"); err == nil {
		execPath = "/usr/bin/chromium-browser"
	} else if _, err := os.Stat("/usr/bin/chromium"); err == nil {
		execPath = "/usr/bin/chromium"
	}

	// Essential Chrome flags for Docker environment
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-features", "TranslateUI"),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.UserAgent(d.cfg.Scraper.UserAgent),
		chromedp.WindowSize(1920, 1080),
	}

	// Добавляем путь к браузеру, если нашли
	if execPath != "" {
		opts = append(opts, chromedp.ExecPath(execPath))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create context with timeout
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set timeout for the whole operation
	taskCtx, cancel = context.WithTimeout(taskCtx, d.cfg.Scraper.Timeout)
	defer cancel()

	var html string

	// Navigation and HTML extraction with better error handling
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Wait for JS to execute
		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		log.Printf("Error downloading page %s: %v", url, err)
		return "", err
	}

	// Сохраняем HTML-файл
	if err := d.SaveHTML(url, html); err != nil {
		log.Printf("Ошибка сохранения HTML: %v", err)
	}

	log.Printf("Successfully downloaded page: %s", url)
	return html, nil
}

// SaveHTML сохраняет HTML страницы в файл
func (d *Downloader) SaveHTML(url, html string) error {
	// Создаем имя файла на основе URL
	filename := d.sanitizeFilename(url) + ".html"
	filepath := filepath.Join(d.cfg.Downloader.OutputDir, "html", filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(html)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	log.Printf("HTML сохранен в: %s", filepath)
	return nil
}

// SaveBlock сохраняет блок в файл
func (d *Downloader) SaveBlock(block *models.Block) error {
	// Создаем директорию для блока
	blockDir := filepath.Join(d.cfg.Downloader.OutputDir, "blocks", block.OperationID.String())
	if err := os.MkdirAll(blockDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории: %w", err)
	}

	// Сохраняем HTML блока
	htmlFilename := fmt.Sprintf("%s_%s.html", block.BlockType, block.ID.String())
	htmlPath := filepath.Join(blockDir, htmlFilename)

	if err := ioutil.WriteFile(htmlPath, []byte(block.HTML), 0644); err != nil {
		return fmt.Errorf("ошибка сохранения HTML блока: %w", err)
	}

	// Сохраняем метаданные блока
	metadata := map[string]interface{}{
		"id":           block.ID,
		"operation_id": block.OperationID,
		"block_type":   block.BlockType,
		"platform":     block.Platform,
		"content":      block.Content,
		"created_at":   block.CreatedAt,
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации метаданных: %w", err)
	}

	metadataFilename := fmt.Sprintf("%s_%s_metadata.json", block.BlockType, block.ID.String())
	metadataPath := filepath.Join(blockDir, metadataFilename)

	if err := ioutil.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("ошибка сохранения метаданных: %w", err)
	}

	log.Printf("Блок сохранен: %s", blockDir)
	return nil
}

// SaveBlocks сохраняет все блоки операции
func (d *Downloader) SaveBlocks(blocks []models.Block) error {
	for _, block := range blocks {
		if err := d.SaveBlock(&block); err != nil {
			log.Printf("Ошибка сохранения блока %s: %v", block.ID, err)
			continue
		}
	}
	return nil
}

// GetBlockFiles возвращает список файлов блоков для операции
func (d *Downloader) GetBlockFiles(operationID uuid.UUID) (map[string][]string, error) {
	blockDir := filepath.Join(d.cfg.Downloader.OutputDir, "blocks", operationID.String())

	// Проверяем существование директории
	if _, err := os.Stat(blockDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("директория блоков не найдена: %s", blockDir)
	}

	files := make(map[string][]string)

	// Проходим по директории
	err := filepath.Walk(blockDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(blockDir, path)
		if err != nil {
			return err
		}

		// Определяем тип файла
		ext := filepath.Ext(rel)
		if ext == ".html" {
			files["html"] = append(files["html"], rel)
		} else if ext == ".json" {
			files["metadata"] = append(files["metadata"], rel)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка чтения директории: %w", err)
	}

	return files, nil
}

// ZipBlocks создает ZIP-архив с блоками операции
func (d *Downloader) ZipBlocks(operationID uuid.UUID) (string, error) {
	blockDir := filepath.Join(d.cfg.Downloader.OutputDir, "blocks", operationID.String())

	// Проверяем существование директории
	if _, err := os.Stat(blockDir); os.IsNotExist(err) {
		return "", fmt.Errorf("директория блоков не найдена: %s", blockDir)
	}

	// Создаем архив
	zipFilename := fmt.Sprintf("blocks_%s.zip", operationID.String())
	zipPath := filepath.Join(d.cfg.Downloader.OutputDir, "exports", zipFilename)

	// Создаем ZIP-файл
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания ZIP-файла: %w", err)
	}
	defer zipFile.Close()

	// Создаем ZIP-writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Добавляем файлы в архив
	err = filepath.Walk(blockDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории
		if info.IsDir() {
			return nil
		}

		// Получаем относительный путь
		relPath, err := filepath.Rel(blockDir, path)
		if err != nil {
			return err
		}

		// Открываем файл
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Создаем файл в архиве
		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Копируем содержимое
		_, err = io.Copy(zipEntry, file)
		if err != nil {
			return err
		}

		log.Printf("Добавлен в архив: %s", relPath)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("ошибка создания архива: %w", err)
	}

	log.Printf("ZIP-архив создан: %s", zipPath)
	return zipPath, nil
}

// DownloadBlock загружает конкретный блок
func (d *Downloader) DownloadBlock(operationID, blockID uuid.UUID, format string) ([]byte, string, error) {
	blockDir := filepath.Join(d.cfg.Downloader.OutputDir, "blocks", operationID.String())

	var filename string
	var data []byte

	switch format {
	case "html":
		// Ищем HTML-файл блока
		files, err := d.GetBlockFiles(operationID)
		if err != nil {
			return nil, "", err
		}

		for _, htmlFile := range files["html"] {
			if strings.Contains(htmlFile, blockID.String()) {
				filepath := filepath.Join(blockDir, htmlFile)
				data, err = ioutil.ReadFile(filepath)
				if err != nil {
					return nil, "", err
				}
				filename = htmlFile
				break
			}
		}

	case "json":
		// Ищем JSON-файл блока
		files, err := d.GetBlockFiles(operationID)
		if err != nil {
			return nil, "", err
		}

		for _, jsonFile := range files["metadata"] {
			if strings.Contains(jsonFile, blockID.String()) {
				filepath := filepath.Join(blockDir, jsonFile)
				data, err = ioutil.ReadFile(filepath)
				if err != nil {
					return nil, "", err
				}
				filename = jsonFile
				break
			}
		}

	default:
		return nil, "", fmt.Errorf("неподдерживаемый формат: %s", format)
	}

	if len(data) == 0 {
		return nil, "", fmt.Errorf("блок не найден")
	}

	return data, filename, nil
}

// sanitizeFilename создает безопасное имя файла из URL
func (d *Downloader) sanitizeFilename(url string) string {
	// Удаляем протокол
	filename := strings.TrimPrefix(url, "http://")
	filename = strings.TrimPrefix(filename, "https://")

	// Заменяем специальные символы
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	// Обрезаем до 255 символов
	if len(filename) > 255 {
		filename = filename[:255]
	}

	return filename
}

// Module регистрирует зависимости для загрузчика
var Module = fx.Module("downloader",
	fx.Provide(
		NewDownloader,
	),
)
