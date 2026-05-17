-- Редактирование сообщений в чате.
ALTER TABLE chat_messages ADD COLUMN edited_at TIMESTAMP WITH TIME ZONE NULL;

-- Скрытие тредов «со своей стороны» (анти-зачистка истории у собеседника).
-- Запись означает: юзер user_id больше не видит тред thread_id у себя в inbox,
-- но сами сообщения остаются в БД и собеседник продолжает видеть тред.
CREATE TABLE chat_thread_hides (
    user_id    UUID NOT NULL,
    thread_id  VARCHAR(128) NOT NULL,
    hidden_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, thread_id)
);
