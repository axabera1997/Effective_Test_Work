CREATE TABLE IF NOT EXISTS subscriptions (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_name VARCHAR(64) NOT NULL,
    price int NOT NULL,
    user_id UUID NOT NULL,
    start_date TIMESTAMP NOT NULL,
    -- если end_date не указано, при размаршалливании в структуру пропишется начальная дата time.Time{}
    end_date TIMESTAMP NOT NULL
);    