-- +goose Up
-- +goose StatementBegin
ALTER TABLE recording DROP COLUMN duration;
ALTER TABLE recording ADD COLUMN ended_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE recording DROP COLUMN ended_at;
ALTER TABLE recording ADD COLUMN duration INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd
