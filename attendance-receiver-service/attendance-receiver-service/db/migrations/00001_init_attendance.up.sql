-- Создание таблицы attendancelog
CREATE TABLE IF NOT EXISTS attendancelog (
    id SERIAL PRIMARY KEY,
    pin TEXT,
    name TEXT,
    last_name TEXT,
    dept_code TEXT,
    dept_name TEXT,
    reader_name TEXT,
    door_name TEXT,
    device_sn TEXT,
    capture_photo_path TEXT,
    verify_mode_name TEXT,
    event_name TEXT,
    event_time TIMESTAMP,
    raw_payload JSONB,
    unique_key TEXT UNIQUE,
    log_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_attendancelog_pin ON attendancelog(pin);
CREATE INDEX IF NOT EXISTS idx_attendancelog_event_time ON attendancelog(event_time);
CREATE INDEX IF NOT EXISTS idx_attendancelog_device_sn ON attendancelog(device_sn);
CREATE INDEX IF NOT EXISTS idx_attendancelog_unique_key ON attendancelog(unique_key);
CREATE INDEX IF NOT EXISTS idx_attendancelog_log_id ON attendancelog(log_id); 