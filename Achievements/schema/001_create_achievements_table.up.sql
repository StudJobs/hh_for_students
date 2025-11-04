-- Создание таблицы достижений
CREATE TABLE IF NOT EXISTS achievements (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_uuid VARCHAR(36) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_achievements_user_uuid ON achievements(user_uuid) WHERE deleted_at IS NULL;
CREATE INDEX idx_achievements_user_name ON achievements(user_uuid, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_achievements_created_at ON achievements(created_at) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_achievements_s3_key ON achievements(s3_key) WHERE deleted_at IS NULL;

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_achievements_updated_at
    BEFORE UPDATE ON achievements
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Комментарии к таблице и колонкам
COMMENT ON TABLE achievements IS 'Таблица для хранения метаданных достижений пользователей';
COMMENT ON COLUMN achievements.name IS 'Уникальное имя достижения в рамках пользователя';
COMMENT ON COLUMN achievements.user_uuid IS 'UUID владельца достижения';
COMMENT ON COLUMN achievements.file_name IS 'Оригинальное имя файла';
COMMENT ON COLUMN achievements.file_type IS 'MIME-тип файла';
COMMENT ON COLUMN achievements.file_size IS 'Размер файла в байтах';
COMMENT ON COLUMN achievements.s3_key IS 'Ключ файла в S3 хранилище';
COMMENT ON COLUMN achievements.created_at IS 'Время создания записи';
COMMENT ON COLUMN achievements.updated_at IS 'Время последнего обновления';
COMMENT ON COLUMN achievements.deleted_at IS 'Время мягкого удаления (NULL если запись активна)';