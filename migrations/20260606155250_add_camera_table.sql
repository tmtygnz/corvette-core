-- +goose Up
-- +goose StatementBegin
CREATE TABLE camera (
    camera_id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_name TEXT NOT NULL,
    installed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL,

    url TEXT NOT NULL,
    sub_url TEXT,
    type TEXT NOT NULL
);

CREATE INDEX idx_camera_status ON camera(status);
CREATE INDEX idx_camera_name ON camera(camera_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_camera_name;
DROP INDEX IF EXISTS idx_camera_status;
DROP TABLE IF EXISTS camera;
-- +goose StatementEnd
