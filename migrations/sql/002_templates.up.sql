-- +goose Up
-- +goose StatementBegin
-- Создаем таблицу для шаблонов блоков
CREATE TABLE IF NOT EXISTS block_templates (
                                               id           SERIAL                     PRIMARY KEY,
                                               block_type   VARCHAR(100)               NOT NULL,
                                               wordpress    JSONB                      NULL,
                                               tilda        JSONB                      NULL,
                                               bitrix       JSONB                      NULL,
                                               html5        JSONB                      NULL,
                                               created_at   TIMESTAMP WITH TIME ZONE   DEFAULT NOW()
);

-- Добавляем базовые шаблоны блоков
INSERT INTO block_templates (block_type, html5) VALUES
                                                    ('Блок с картинкой 3 колонки',       '{"priority": 4, "step1": "<img",             "step2": "col-3"}'),
                                                    ('Карта',                            '{"priority": 0, "step1": "map"}'),
                                                    ('Картинка + Действие',              '{"priority": 4, "step1": "<img",             "step2": "<button|<a"}'),
                                                    ('Картинка+текст',                   '{"priority": 5, "step1": "<img",             "step2": "<p|<h"}'),
                                                    ('Карточка товара',                  '{"priority": 0, "step1": "tovar"}'),
                                                    ('Контакты',                         '{"priority": 0, "step1": "contacts"}'),
                                                    ('Партнеры',                         '{"priority": 0, "step1": "partners"}'),
                                                    ('Поиск',                            '{"priority": 0, "step1": "find"}'),
                                                    ('Смешанный контент',                '{"priority": 3, "step1": "<p|<h",            "step2": "<button|<a"}'),
                                                    ('Таблица',                          '{"priority": 0, "step1": "table"}'),
                                                    ('Таймлайн',                         '{"priority": 1, "step1": "timeline"}'),
                                                    ('Текст + Действие',                 '{"priority": 5, "step1": "<p|<h",            "step2": "<button|<a"}'),
                                                    ('Текстовый блок',                   '{"priority": 4, "step1": "<p|<h"}'),
                                                    ('Текстовый блок 2 колонки',         '{"priority": 5, "step1": "<p|<h",            "step2": "col-2"}'),
                                                    ('Попап, виджет',                    '{"priority": 5, "step1": "popup|onclick"}'),
                                                    ('Текст блок + Картинка',            '{"priority": 4, "step1": "<p|<h",            "step2": "<img"}'),
                                                    ('Товары',                           '{"priority": 1, "step1": "products"}'),
                                                    ('Карусель, слайд шоу с текстом',    '{"priority": 0, "step1": "swiper",           "step2": "<p|<h"}'),
                                                    ('FAQ',                              '{"priority": 1, "step1": "faq"}'),
                                                    ('Форма обратной связи',             '{"priority": 1, "step1": "mailto:tel:"}'),
                                                    ('Карусель, слайд шоу',              '{"priority": 1, "step1": "swiper"}'),
                                                    ('Блок с картинкой',                 '{"priority": 5, "step1": "<img"}');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Удаление таблицы block_templates
DROP TABLE IF EXISTS block_templates CASCADE;
-- +goose StatementEnd
