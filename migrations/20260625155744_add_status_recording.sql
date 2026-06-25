-- +goose Up
-- +goose StatementBegin
ALTER TABLE recording ADD COLUMN status TEXT NOT NULL DEFAULT 'done';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE recording DROP COLUMN status;
-- +goose StatementEnd
