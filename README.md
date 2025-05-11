# Website Scraper

Поисковый робот для сбора данных с сайтов, который автоматически находит и классифицирует стандартные блоки сайтов (шапки, подвалы и т.д.) различных платформ (WordPress, Tilda, Bitrix, HTML5).

## Авторы

- Разуваев Денис Русланович
- Третьяков Алексей Альбертович
- Галимова Динара Фалыховна
- Кулаков Арсений Вячеславович

## Описание проекта

Website Scraper - это веб-сервис для парсинга и анализа структуры сайтов. Он автоматически определяет платформу сайта, находит основные блоки (шапка, подвал и др.) и предоставляет возможность скачать результаты в различных форматах.

### Основные возможности

- Автоматическое определение платформы сайта (WordPress, Tilda, Bitrix, HTML5)
- Распознавание и извлечение шапок и подвалов сайтов
- Классификация контентных блоков
- Сохранение и экспорт результатов анализа
- API для автоматизации процесса парсинга
- Создание сводного отчета по всем найденным блокам

### Поддерживаемые платформы

- WordPress
- Tilda
- Bitrix
- HTML5

## Требования

- Docker и Docker Compose
- Доступ к интернету для парсинга сайтов

## Установка и запуск

### С использованием Docker Compose

1. Клонировать репозиторий:
```bash
git clone https://github.com/yourusername/website-scraper.git
cd website-scraper
```

2. Запустить сервисы с помощью Docker Compose:
```bash
docker-compose up -d
```

3. Сервис будет доступен по адресу http://localhost:8080

### Сборка и запуск вручную

1. Установить Go (версия 1.16 или выше)
2. Клонировать репозиторий
3. Установить зависимости:
```bash
go mod download
```
4. Собрать проект:
```bash
go build -o app ./cmd/app
```
5. Запустить:
```bash
./app
```

## Примеры использования

### API методы

Сервис предоставляет следующие API эндпоинты:

#### Парсинг URL

```bash
curl -X POST http://localhost:8080/api/v1/parse \
  -H "Content-Type: application/json" \
  -d '{"url": "https://structura.app"}'
```

Пример ответа:
```json
{
  "operation_id": "e145e890-4d66-4310-b94a-fa0ebef513be",
  "links": {
    "blocks_list": "/api/v1/operations/e145e890-4d66-4310-b94a-fa0ebef513be/blocks",
    "download": "/api/v1/download/e145e890-4d66-4310-b94a-fa0ebef513be",
    "export": "/api/v1/operations/e145e890-4d66-4310-b94a-fa0ebef513be/export",
    "get_result": "/api/v1/operations/e145e890-4d66-4310-b94a-fa0ebef513be",
    "save_blocks": "/api/v1/operations/e145e890-4d66-4310-b94a-fa0ebef513be/blocks/save"
  },
  "message": "Операция парсинга запущена. Используйте предоставленные ссылки для получения результатов и экспорта."
}
```

#### Получение результатов операции

```bash
curl -X GET http://localhost:8080/api/v1/operations/{operation_id}
```

#### Экспорт результатов операции

```bash
# Экспорт в Excel
curl -X GET http://localhost:8080/api/v1/operations/{operation_id}/export?format=excel -o results.xlsx

# Экспорт в текстовый файл
curl -X GET http://localhost:8080/api/v1/operations/{operation_id}/export?format=text -o results.txt
```

#### Сохранение блоков операции

```bash
curl -X POST http://localhost:8080/api/v1/operations/{operation_id}/blocks/save
```

#### Получение списка файлов блоков

```bash
curl -X GET http://localhost:8080/api/v1/operations/{operation_id}/blocks
```

#### Скачивание конкретного блока

```bash
# Скачивание блока в формате HTML
curl -X GET http://localhost:8080/api/v1/operations/{operation_id}/blocks/{block_id}/download?format=html -o block.html

# Скачивание блока в формате JSON
curl -X GET http://localhost:8080/api/v1/operations/{operation_id}/blocks/{block_id}/download?format=json -o block.json
```

#### Скачивание всех блоков операции в ZIP-архиве

```bash
curl -X GET http://localhost:8080/api/v1/operations/{operation_id}/blocks/download -o blocks.zip
```

#### Обход URL и сбор ссылок

```bash
curl -X POST http://localhost:8080/api/v1/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://structura.app",
    "max_depth": 2,
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
  }'
```

### Полный тестовый сценарий

Ниже приведен скрипт для тестирования всех основных функций системы:

```bash
#!/bin/bash

echo "=== ТЕСТИРОВАНИЕ API ПАРСЕРА ==="
echo "================================="

# 1. Парсинг URL
echo "1. Создание операции парсинга..."
PARSE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/parse \
  -H "Content-Type: application/json" \
  -d '{"url": "https://structura.app"}')
echo "Ответ:"
echo "$PARSE_RESPONSE"

# Извлекаем ID операции
OPERATION_ID=$(echo $PARSE_RESPONSE | grep -o '"operation_id":"[^"]*"' | cut -d'"' -f4)
echo "ID операции: $OPERATION_ID"
echo ""

# 2. Ожидание завершения операции
echo "2. Ожидание завершения операции (30 секунд)..."
sleep 30

# 3. Получение результатов операции
echo "3. Проверка результатов операции..."
curl -s -X GET http://localhost:8080/api/v1/operations/$OPERATION_ID
echo ""

# 4. Сохранение блоков
echo "4. Сохранение блоков операции..."
curl -s -X POST http://localhost:8080/api/v1/operations/$OPERATION_ID/blocks/save
echo ""

# 5. Получение списка файлов блоков
echo "5. Получение списка файлов блоков..."
curl -s -X GET http://localhost:8080/api/v1/operations/$OPERATION_ID/blocks
echo ""

# 6. Скачивание блоков как ZIP
echo "6. Скачивание архива со всеми блоками..."
curl -s -X GET http://localhost:8080/api/v1/operations/$OPERATION_ID/blocks/download -o blocks_$OPERATION_ID.zip
echo "Архив сохранен как: blocks_$OPERATION_ID.zip"
echo ""

# 7. Экспорт в Excel
echo "7. Экспорт результатов в Excel..."
curl -s -X GET http://localhost:8080/api/v1/operations/$OPERATION_ID/export?format=excel -o operation_$OPERATION_ID.xlsx
echo "Файл сохранен как: operation_$OPERATION_ID.xlsx"
echo ""

# 8. Экспорт в текст
echo "8. Экспорт результатов в текстовый файл..."
curl -s -X GET http://localhost:8080/api/v1/operations/$OPERATION_ID/export?format=text -o operation_$OPERATION_ID.txt
echo "Файл сохранен как: operation_$OPERATION_ID.txt"
echo ""

# 9. Проверка скачанных файлов
echo "9. Проверка скачанных файлов..."
XLSX_SIZE=$(stat -c%s "operation_$OPERATION_ID.xlsx" 2>/dev/null || echo "0")
TXT_SIZE=$(stat -c%s "operation_$OPERATION_ID.txt" 2>/dev/null || echo "0")
ZIP_SIZE=$(stat -c%s "blocks_$OPERATION_ID.zip" 2>/dev/null || echo "0")

echo "Excel: $XLSX_SIZE bytes"
echo "Text: $TXT_SIZE bytes"
echo "ZIP:  $ZIP_SIZE bytes"
echo ""

# 10. Обход URL и сбор ссылок
echo "10. Тестирование краулера..."
curl -s -X POST http://localhost:8080/api/v1/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://structura.app",
    "max_depth": 1
  }'
echo ""

echo "=== ТЕСТ ЗАВЕРШЕН ==="
```

## Архитектура проекта

Website Scraper использует модульную архитектуру с разделением ответственности между компонентами:

- **API Layer** - Обработка HTTP-запросов и ответов
- **Parser Layer** - Анализ и извлечение блоков из HTML
- **Platform Detectors** - Определение платформы сайта
- **Block Parsers** - Специализированные парсеры для разных типов блоков
- **Crawler** - Обход ссылок на сайте
- **Downloader** - Сохранение и экспорт результатов
- **Repository Layer** - Работа с хранилищем данных

