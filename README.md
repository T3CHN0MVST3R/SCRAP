# Website Scraper API

Сервис для автоматического парсинга и анализа веб-сайтов различных платформ с высокой точностью (>70%).

## 🚀 Особенности

- **Поддержка различных платформ**: WordPress, Tilda, Битрикс, HTML5
- **Автоматическое определение шапок и подвалов** сайтов
- **Гибкая классификация контентных блоков** с использованием шаблонов
- **Краулинг с настраиваемой глубиной** и параллелизмом
- **Экспорт результатов** в Excel и текстовый формат
- **RESTful API** с автоматической документацией Swagger
- **Docker-based** архитектура для простого развертывания

## 🛠 Технологический стек

- **Backend**: Go 1.24, Fx (Dependency Injection)
- **HTTP**: Gorilla Mux, Swagger/OpenAPI
- **Database**: PostgreSQL, Goose для миграций
- **Browser Engine**: ChromeDP (headless Chrome)
- **Контейнеризация**: Docker, Docker Compose
- **Parsing**: GoQuery, regexp для анализа HTML

## 📋 Требования

- Docker и Docker Compose
- Git
- Порт 8080 (приложение) и 5434 (PostgreSQL)
- Минимум 1GB RAM

## 🚀 Быстрый старт

### 1. Клонирование репозитория

```bash
git clone https://github.com/T3CHN0MVST3R/SCRAPPER.git
cd SCRAPPER
```

### 2. Запуск с Docker Compose

```bash
docker compose up --build
```

Приложение будет доступно по адресу: `http://localhost:8080`

## 📚 API Документация

### Основные эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/parse` | Запустить парсинг URL |
| `GET` | `/api/v1/operations/{id}` | Получить результаты операции |
| `GET` | `/api/v1/operations/{id}/export` | Экспортировать результаты |
| `GET` | `/api/v1/download/{id}` | Скачать результаты |
| `POST` | `/api/v1/crawl` | Запустить краулинг |
| `GET` | `/api/v1/formats` | Получить доступные форматы |


## 💡 Примеры использования

### Парсинг сайта

```bash
curl -X POST http://localhost:8080/api/v1/parse \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://structura.app"
  }'
```

### Получение результатов

```bash
curl -X GET http://localhost:8080/api/v1/operations/OPERATION_ID
```

### Краулинг с настройками

```bash
curl -X POST http://localhost:8080/api/v1/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "max_depth": 2,
    "user_agent": "Custom Bot 1.0"
  }'
```

### Экспорт в Excel

```bash
curl -X GET "http://localhost:8080/api/v1/operations/OPERATION_ID/export?format=excel" \
  -o results.xlsx
```

## 🔧 Конфигурация

Основные переменные окружения:

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `SERVER_PORT` | Порт приложения | `8080` |
| `DB_HOST` | Хост базы данных | `postgres` |
| `DB_PORT` | Порт базы данных | `5432` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | `postgres` |
| `DB_NAME` | Имя базы данных | `scraper` |
| `SCRAPER_TIMEOUT` | Таймаут запросов | `30s` |
| `SCRAPER_MAX_DEPTH` | Максимальная глубина краулинга | `2` |

## 🎯 Поддерживаемые блоки

### Шапки (Headers)
- WordPress шапка
- Tilda шапка  
- Битрикс шапка
- HTML5 шапка

### Подвалы (Footers)
- WordPress подвал
- Tilda подвал
- Битрикс подвал
- HTML5 подвал

### Контентные блоки
- Текстовые блоки
- Изображения
- Формы обратной связи
- Карты
- Карусели / слайдеры
- FAQ секции
- И многое другое...

## 🔍 Примеры тестовых сайтов

- `https://botcreators.ru` (WordPress)
- `https://structura.app` (Tilda)
- `https://automatisation.art` (WordPress)
- `https://mindbox.ru` (HTML5)
- `https://skillfactory.ru` (HTML5)

## 📊 Тестирование

Для проверки работы используйте команды из файла `curl_commands.md`:

```bash
# Проверка статуса сервиса
curl http://localhost:8080/

# Запуск парсинга
curl -X POST http://localhost:8080/api/v1/parse \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для вашей функции (`git checkout -b feature/amazing-feature`)
3. Сделайте commit (`git commit -m 'Add some amazing feature'`)
4. Пушните в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

