package app

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/gorilla/mux"

	"website-scraper/internal/config"
	"website-scraper/migrations"
)

// Module регистрирует зависимости для приложения
var Module = fx.Module("app",
	fx.Invoke(
		RunMigrations,
		StartServer,
	),
)

// RunMigrations выполняет миграции базы данных
func RunMigrations(db *sql.DB) error {
	return migrations.RunMigrations(db)
}

// StartServer запускает HTTP-сервер
func StartServer(lifecycle fx.Lifecycle, router *mux.Router, cfg *config.Config) {
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					panic(err)
				}
			}()

			// Обработка сигналов для грациозного завершения
			go func() {
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
				<-sigChan

				// Создаем контекст с таймаутом для завершения
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := server.Shutdown(shutdownCtx); err != nil {
					panic(err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
