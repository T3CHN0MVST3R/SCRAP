package routes

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	"website-scraper/internal/api/handlers"
)

// SetupRouter создает и настраивает маршрутизатор
func SetupRouter(handlers *handlers.Handlers) *mux.Router {
	router := mux.NewRouter()

	// API маршруты
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Регистрируем маршруты парсера
	apiRouter.HandleFunc("/parse", handlers.ParseURL).Methods(http.MethodPost)
	apiRouter.HandleFunc("/operations/{id}", handlers.GetOperationResult).Methods(http.MethodGet)
	apiRouter.HandleFunc("/operations/{id}/export", handlers.ExportOperation).Methods(http.MethodGet)

	// Регистрируем маршруты загрузчика
	apiRouter.HandleFunc("/download/{id}", handlers.DownloadByID).Methods(http.MethodGet)
	apiRouter.HandleFunc("/formats", handlers.GetFormats).Methods(http.MethodGet)

	// Регистрируем маршруты краулера
	apiRouter.HandleFunc("/crawl", handlers.CrawlURL).Methods(http.MethodPost)

	// Добавьте эти строки в функцию SetupRouter
	apiRouter.HandleFunc("/operations/{operation_id}/blocks/save", handlers.SaveBlocksEndpoint).Methods(http.MethodPost)
	apiRouter.HandleFunc("/operations/{operation_id}/blocks", handlers.GetBlockFiles).Methods(http.MethodGet)
	apiRouter.HandleFunc("/operations/{operation_id}/blocks/{block_id}/download", handlers.DownloadBlock).Methods(http.MethodGet)
	apiRouter.HandleFunc("/operations/{operation_id}/blocks/download", handlers.DownloadAllBlocks).Methods(http.MethodGet)

	// Swagger UI
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Простая главная страница
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
		<!DOCTYPE html>
		<html>
			<head>
				<title>Website Scraper API</title>
				<style>
					body { font-family: Arial, sans-serif; line-height: 1.6; max-width: 800px; margin: 0 auto; padding: 20px; }
					h1 { color: #333; }
					ul { list-style-type: none; padding: 0; }
					li { margin-bottom: 10px; }
					a { color: #0066cc; text-decoration: none; }
					a:hover { text-decoration: underline; }
					.endpoint { background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 15px; }
					.method { display: inline-block; padding: 5px 10px; border-radius: 3px; color: white; font-weight: bold; margin-right: 10px; }
					.get { background-color: #61affe; }
					.post { background-color: #49cc90; }
					.endpoint-url { font-family: monospace; }
				</style>
			</head>
			<body>
				<h1>Website Scraper API</h1>
				<p>Сервис для парсинга и анализа веб-сайтов.</p>
				
				<h2>Доступные эндпоинты:</h2>
				
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="endpoint-url">/api/v1/parse</span>
					<p>Парсит указанный URL и возвращает ID операции.</p>
				</div>
				
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="endpoint-url">/api/v1/operations/{id}</span>
					<p>Возвращает результаты операции парсинга по ID.</p>
				</div>
				
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="endpoint-url">/api/v1/operations/{id}/export</span>
					<p>Экспортирует результаты операции в указанном формате.</p>
				</div>
				
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="endpoint-url">/api/v1/download/{id}</span>
					<p>Загружает файлы по ID операции в указанном формате.</p>
				</div>
				
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="endpoint-url">/api/v1/formats</span>
					<p>Возвращает список доступных форматов для загрузки.</p>
				</div>
				
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="endpoint-url">/api/v1/crawl</span>
					<p>Обходит указанный URL и собирает ссылки.</p>
				</div>
				
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="endpoint-url">/api/v1/operations/{operation_id}/blocks/save</span>
					<p>Сохраняет все блоки операции в файлы.</p>
				</div>
				
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="endpoint-url">/api/v1/operations/{operation_id}/blocks</span>
					<p>Возвращает список файлов блоков для операции.</p>
				</div>
				
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="endpoint-url">/api/v1/operations/{operation_id}/blocks/{block_id}/download</span>
					<p>Загружает конкретный блок в выбранном формате.</p>
				</div>
				
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="endpoint-url">/api/v1/operations/{operation_id}/blocks/download</span>
					<p>Создает и загружает архив со всеми блоками операции.</p>
				</div>
				
				<p>Документация API доступна по ссылке: <a href="/swagger/">/swagger/</a></p>
			</body>
		</html>
		`))
	})

	// Обработчик ошибок 404
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Эндпоинт не найден"}`))
	})

	return router
}

// Module регистрирует зависимости для маршрутов
var Module = fx.Module("routes",
	fx.Provide(
		handlers.NewHandlers,
		SetupRouter,
	),
)
