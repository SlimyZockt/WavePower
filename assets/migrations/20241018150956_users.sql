-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY  NOT NULL,
    email VARCHAR(255) NOT NULL,
    playlist TEXT NOT NULL,
    name VARCHAR(255) NOT NULL

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP Table users;
-- +goose StatementEnd
