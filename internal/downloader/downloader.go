package downloader

import (
	"context"
	"time"

	"go.uber.org/fx"

	"github.com/chromedp/chromedp"

	"website-scraper/internal/config"
)

// Downloader представляет сервис для загрузки веб-страниц
type Downloader struct {
	cfg *config.Config
}

// NewDownloader создает новый экземпляр Downloader
func NewDownloader(cfg *config.Config) *Downloader {
	return &Downloader{
		cfg: cfg,
	}
}

// DownloadPage загружает страницу и возвращает HTML
func (d *Downloader) DownloadPage(ctx context.Context, url string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent(d.cfg.Scraper.UserAgent),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Создаем контекст с таймаутом
	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Устанавливаем общий таймаут на выполнение
	taskCtx, cancel = context.WithTimeout(taskCtx, d.cfg.Scraper.Timeout)
	defer cancel()

	var html string

	// Навигация и извлечение HTML
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second), // Дать JS отработать
		chromedp.OuterHTML("html", &html),
	)

	return html, err
}

// Module регистрирует зависимости для загрузчика
var Module = fx.Module("downloader",
	fx.Provide(
		NewDownloader,
	),
)
