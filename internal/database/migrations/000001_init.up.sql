-- Create the main table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL CONSTRAINT users_email_unique UNIQUE,
    password_hash TEXT NOT NULL
)
