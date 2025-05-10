package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/fx"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"website-scraper/internal/config"
)

// CrawlerService интерфейс для сервиса краулера
type CrawlerService interface {
	// CrawlURL обходит URL и собирает ссылки
	CrawlURL(ctx context.Context, url string, maxDepth int) ([]string, error)

	// IsAllowedDomain проверяет, разрешен ли домен для обхода
	IsAllowedDomain(url string) bool

	// SetUserAgent устанавливает User-Agent для запросов
	SetUserAgent(userAgent string)

	// SetMaxDepth устанавливает максимальную глубину обхода
	SetMaxDepth(depth int)
}

// crawlerService реализация CrawlerService
type crawlerService struct {
	config          *config.Config
	userAgent       string
	maxDepth        int
	allowedDomains  []string
	client          *http.Client
	visitedURLs     sync.Map
	concurrency     int
	crawlDelay      time.Duration
	domainLastVisit sync.Map // Для отслеживания времени последнего посещения домена
}

// NewCrawlerService создает новый экземпляр CrawlerService
func NewCrawlerService(cfg *config.Config) CrawlerService {
	return &crawlerService{
		config:         cfg,
		userAgent:      cfg.Scraper.UserAgent,
		maxDepth:       cfg.Scraper.MaxDepth,
		allowedDomains: cfg.Scraper.AllowedDomains,
		client: &http.Client{
			Timeout: cfg.Scraper.Timeout,
		},
		visitedURLs:     sync.Map{},
		domainLastVisit: sync.Map{},
		concurrency:     cfg.Scraper.Concurrency,
		crawlDelay:      cfg.Scraper.CrawlDelay,
	}
}

// CrawlURL обходит URL и собирает ссылки
func (s *crawlerService) CrawlURL(ctx context.Context, urlStr string, maxDepth int) ([]string, error) {
	// Парсим начальный URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse URL")
	}

	// Проверяем, что домен разрешен
	if !s.IsAllowedDomain(urlStr) {
		return nil, fmt.Errorf("domain not allowed: %s", parsedURL.Host)
	}

	// Устанавливаем максимальную глубину
	if maxDepth > 0 {
		s.maxDepth = maxDepth
	}

	// Создаем канал для результатов
	results := make(chan string, 100)
	resultList := []string{}

	// Создаем WaitGroup для горутин обхода
	var wg sync.WaitGroup

	// Создаем семафор для ограничения параллелизма
	sem := make(chan struct{}, s.concurrency)

	// Начинаем с начального URL
	wg.Add(1)
	go func() {
		s.crawlRecursive(ctx, urlStr, 0, results, &wg, sem)
	}()

	// Создаем горутину для закрытия канала результатов когда обход закончен
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем результаты из канала
	for url := range results {
		resultList = append(resultList, url)
	}

	return resultList, nil
}

// crawlRecursive рекурсивно обходит URLs до указанной глубины
func (s *crawlerService) crawlRecursive(
	ctx context.Context,
	urlStr string,
	depth int,
	results chan<- string,
	wg *sync.WaitGroup,
	sem chan struct{},
) {
	defer wg.Done()

	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Проверяем ограничение глубины
	if depth > s.maxDepth {
		return
	}

	// Нормализуем URL
	normalizedURL, err := s.normalizeURL(urlStr)
	if err != nil {
		return
	}

	// Проверяем, был ли URL уже посещен
	if _, visited := s.visitedURLs.LoadOrStore(normalizedURL, true); visited {
		return
	}

	// Получаем семафор (ограничиваем параллелизм)
	sem <- struct{}{}
	defer func() { <-sem }()

	// Применяем задержку для домена
	domainKey := getDomainKey(normalizedURL)
	s.applyDomainRateLimit(domainKey)

	// Отправляем запрос с User-Agent
	req, err := http.NewRequestWithContext(ctx, "GET", normalizedURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", s.userAgent)

	// Выполняем запрос
	resp, err := s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return
	}

	// Добавляем URL в результаты
	results <- normalizedURL

	// Парсим HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	// Извлекаем ссылки
	links := s.extractLinks(doc, normalizedURL)

	// Рекурсивно обходим ссылки
	for _, link := range links {
		if s.IsAllowedDomain(link) {
			wg.Add(1)
			go func(url string) {
				s.crawlRecursive(ctx, url, depth+1, results, wg, sem)
			}(link)
		}
	}
}

// applyDomainRateLimit применяет ограничение скорости для конкретного домена
func (s *crawlerService) applyDomainRateLimit(domainKey string) {
	now := time.Now()

	// Получаем время последнего посещения домена
	if lastVisitObj, exists := s.domainLastVisit.Load(domainKey); exists {
		lastVisit := lastVisitObj.(time.Time)
		// Вычисляем, сколько времени прошло с последнего посещения
		elapsed := now.Sub(lastVisit)

		// Если прошло меньше времени, чем crawlDelay, ждем оставшееся время
		if elapsed < s.crawlDelay {
			waitTime := s.crawlDelay - elapsed
			time.Sleep(waitTime)
		}
	}

	// Обновляем время последнего посещения
	s.domainLastVisit.Store(domainKey, time.Now())
}

// normalizeURL нормализует URL
func (s *crawlerService) normalizeURL(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Удаляем фрагмент
	parsedURL.Fragment = ""

	// Обеспечиваем наличие схемы
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	// Удаляем конечный слеш для согласованности в дедупликации
	if parsedURL.Path == "/" {
		parsedURL.Path = ""
	}

	return parsedURL.String(), nil
}

// extractLinks извлекает ссылки из HTML документа
func (s *crawlerService) extractLinks(doc *goquery.Document, baseURLStr string) []string {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil
	}

	var links []string
	doc.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			return
		}

		// Парсим ссылку
		linkURL, err := url.Parse(href)
		if err != nil {
			return
		}

		// Разрешаем относительные URLs
		resolvedURL := baseURL.ResolveReference(linkURL)

		// Пропускаем не-HTTP(S) URLs
		if resolvedURL.Scheme != "http" && resolvedURL.Scheme != "https" {
			return
		}

		// Удаляем запрос и фрагмент
		resolvedURL.RawQuery = ""
		resolvedURL.Fragment = ""

		// Добавляем ссылку в результаты
		links = append(links, resolvedURL.String())
	})

	return links
}

// IsAllowedDomain проверяет, разрешен ли домен для обхода
func (s *crawlerService) IsAllowedDomain(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	host := parsedURL.Hostname()

	// Проверяем, есть ли домен в списке разрешенных
	for _, allowedDomain := range s.allowedDomains {
		if host == allowedDomain || strings.HasSuffix(host, "."+allowedDomain) {
			return true
		}
	}

	return false
}

// SetUserAgent устанавливает User-Agent для запросов
func (s *crawlerService) SetUserAgent(userAgent string) {
	s.userAgent = userAgent
}

// SetMaxDepth устанавливает максимальную глубину обхода
func (s *crawlerService) SetMaxDepth(depth int) {
	if depth < 1 {
		s.maxDepth = 1
	} else {
		s.maxDepth = depth
	}
}

// SetConcurrency устанавливает максимальное количество параллельных запросов
func (s *crawlerService) SetConcurrency(concurrency int) {
	if concurrency < 1 {
		concurrency = 1
	}
	s.concurrency = concurrency
}

// SetCrawlDelay устанавливает задержку между запросами к одному домену
func (s *crawlerService) SetCrawlDelay(delay time.Duration) {
	if delay < 0 {
		delay = 0
	}
	s.crawlDelay = delay
}

// getDomainKey возвращает ключ домена из URL
func getDomainKey(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	return parsedURL.Hostname()
}

// Module регистрирует зависимости для краулера
var Module = fx.Module("crawler",
	fx.Provide(
		NewCrawlerService,
	),
)
