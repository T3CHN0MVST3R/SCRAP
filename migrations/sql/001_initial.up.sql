-- +goose Up
-- +goose StatementBegin

-- Подключаем расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица для хранения операций парсинга
CREATE TABLE IF NOT EXISTS operations (
                                          id          UUID                   PRIMARY KEY DEFAULT uuid_generate_v4(),
                                          url         TEXT                   NOT NULL,
                                          status      VARCHAR(20)            NOT NULL
                                              CHECK (status IN ('pending', 'processing', 'completed', 'error')),
                                          created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                          updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для ускорения выборок в operations
CREATE INDEX IF NOT EXISTS idx_operations_status     ON operations(status);
CREATE INDEX IF NOT EXISTS idx_operations_created_at ON operations(created_at);

-- Таблица для хранения блоков, найденных при парсинге
CREATE TABLE IF NOT EXISTS blocks (
                                      id           UUID                   PRIMARY KEY DEFAULT uuid_generate_v4(),
                                      operation_id UUID                   NOT NULL
                                          REFERENCES operations(id) ON DELETE CASCADE,
                                      block_type   VARCHAR(20)            NOT NULL
                                          CHECK (block_type IN ('header', 'footer', 'content')),
                                      platform     VARCHAR(20)            NOT NULL
                                          CHECK (platform IN ('wordpress', 'tilda', 'bitrix', 'html5', 'unknown')),
                                      content      JSONB                  NOT NULL,
                                      html         TEXT                   NOT NULL,
                                      created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для ускорения выборок в blocks
CREATE INDEX IF NOT EXISTS idx_blocks_operation_id ON blocks(operation_id);
CREATE INDEX IF NOT EXISTS idx_blocks_block_type   ON blocks(block_type);
CREATE INDEX IF NOT EXISTS idx_blocks_platform     ON blocks(platform);

-- Таблица для хранения ссылок, найденных краулером
CREATE TABLE IF NOT EXISTS links (
                                     id           UUID                   PRIMARY KEY DEFAULT uuid_generate_v4(),
                                     operation_id UUID                   NOT NULL
                                         REFERENCES operations(id) ON DELETE CASCADE,
                                     url          TEXT                   NOT NULL,
                                     status       INT                    DEFAULT 200,
                                     depth        INT                    NOT NULL DEFAULT 0,
                                     created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для ускорения выборок в links
CREATE INDEX  IF NOT EXISTS idx_links_operation_id  ON links(operation_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_links_operation_url ON links(operation_id, url);

-- Функция для авто-обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для operations: обновление updated_at
CREATE TRIGGER update_operations_updated_at
    BEFORE UPDATE ON operations
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS update_operations_updated_at ON operations;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем таблицы в обратном порядке
DROP TABLE IF EXISTS links       CASCADE;
DROP TABLE IF EXISTS blocks      CASCADE;
DROP TABLE IF EXISTS operations  CASCADE;

-- +goose StatementEnd
