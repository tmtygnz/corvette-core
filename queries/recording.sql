-- name: CreateRecording :one
INSERT INTO recording (
    from_camera,
    file_name,
    started_at
)
VALUES (?, ?, ?)
RETURNING *;

-- name: SetEndTime :one
UPDATE recording
SET ended_at = ?
WHERE record_id = ?
RETURNING *;

-- name: GetRecordingFor :many
SELECT *
FROM recording
WHERE from_camera = ?
  AND started_at < ?
  AND (ended_at IS NULL OR ended_at >= ?)
ORDER BY started_at ASC;

-- name: ListRecordings :many
SELECT *
FROM recording
ORDER BY started_at ASC;

-- name: GetRecordingByID :one
SELECT *
FROM recording
WHERE record_id = ?;

-- name: DeleteRecording :exec
DELETE FROM recording
WHERE record_id = ?;
