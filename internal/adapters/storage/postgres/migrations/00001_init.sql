-- +goose Up
-- Users table
CREATE TABLE IF NOT EXISTS users
(
    id            SERIAL PRIMARY KEY,
    login         VARCHAR(255)                NOT NULL UNIQUE,
    password_hash VARCHAR(255)                NOT NULL,
    created_at    timestamp without time zone NOT NULL DEFAULT (current_timestamp AT TIME ZONE 'UTC')
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders
(
    id         SERIAL PRIMARY KEY,
    user_id    INT                         NOT NULL REFERENCES users (id),
    number     VARCHAR                     NOT NULL UNIQUE,
    status     VARCHAR,
    accrual    BIGINT,
    withdraw   BIGINT,
    created_at timestamp without time zone NOT NULL DEFAULT (current_timestamp AT TIME ZONE 'UTC'),
    updated_at timestamp without time zone          DEFAULT (current_timestamp AT TIME ZONE 'UTC')
);
-- +goose Down
DROP TABLE orders;
DROP TABLE users;
