-- +goose Up
-- +goose StatementBegin
CREATE TABLE recording (
    record_id   INTEGER PRIMARY KEY,
    from_camera INTEGER NOT NULL REFERENCES camera(camera_id),
    file_name   TEXT NOT NULL,
    started_at  DATETIME NOT NULL,
    duration    INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE recording;
-- +goose StatementEnd
