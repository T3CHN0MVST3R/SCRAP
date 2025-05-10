package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Scraper    ScraperConfig
	Downloader DownloaderConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

type ScraperConfig struct {
	UserAgent      string
	Timeout        time.Duration
	MaxDepth       int
	AllowedDomains []string
	Concurrency    int
	CrawlDelay     time.Duration
}

type DownloaderConfig struct {
	OutputDir string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultValue
}

func New() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnv("DB_PORT", "5432"),
			Username: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "scraper"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Scraper: ScraperConfig{
			UserAgent:   getEnv("SCRAPER_USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"),
			Timeout:     getEnvDuration("SCRAPER_TIMEOUT", 30*time.Second),
			MaxDepth:    getEnvInt("SCRAPER_MAX_DEPTH", 2),
			Concurrency: getEnvInt("SCRAPER_CONCURRENCY", 5),
			CrawlDelay:  getEnvDuration("SCRAPER_CRAWL_DELAY", 1*time.Second),
			AllowedDomains: []string{
				"botcreators.ru",
				"structura.app",
				"automatisation.art",
				"mindbox.ru",
				"skillfactory.ru",
			},
		},
		Downloader: DownloaderConfig{
			OutputDir: getEnv("DOWNLOADER_OUTPUT_DIR", "./downloads"),
		},
	}
}

// GetPostgresDSN возвращает строку подключения к PostgreSQL
func (c *DatabaseConfig) GetPostgresDSN() string {
	return "postgres://" + c.Username + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DBName + "?sslmode=" + c.SSLMode
}
