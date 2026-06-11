-- name: CreateRecording :one
INSERT INTO recording (
    from_camera,
    file_name,
    started_at,
    duration
)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: SetDuration :one
UPDATE recording
SET duration = ?
WHERE record_id = ?
RETURNING *;

-- name: ListRecordings :many
SELECT * FROM recording ORDER BY started_at;

-- name: DeleteRecording :exec
DELETE FROM recording WHERE record_id = ?;
