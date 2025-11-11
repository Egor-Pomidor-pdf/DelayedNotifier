CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    recipient TEXT NOT NULL,          -- email, telegram id и т.д.
    channel TEXT NOT NULL,            -- email, telegram
    message TEXT NOT NULL,            -- текст уведомления
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,-- время отправки
    status TEXT NOT NULL DEFAULT 'pending',  -- pending / sent / cancelled / failed
    tries INT NOT NULL DEFAULT 0,    -- количество попыток отправки
    last_error TEXT,                  -- текст последней ошибки
    created_at     TIMESTAMP WITH TIME ZONE          DEFAULT now(),
    updated_at     TIMESTAMP WITH TIME ZONE          DEFAULT now()
);