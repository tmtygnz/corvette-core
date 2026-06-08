-- name: CreateCamera :one
INSERT INTO camera (
    camera_name,
    installed_at,
    status,
    url,
    sub_url,
    type
)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetCamera :one
SELECT *
FROM camera
WHERE camera_id = ?
LIMIT 1;

-- name: ListCameras :many
SELECT *
FROM camera
ORDER BY camera_id;

-- name: UpdateCamera :one
UPDATE camera
SET
    camera_name = ?,
    url = ?,
    sub_url = ?,
    type = ?
WHERE camera_id = ?
RETURNING *;

-- name: UpdateCameraStatus :one
UPDATE camera
SET status = ?
WHERE camera_id = ?
RETURNING *;

-- name: DeleteCamera :exec
DELETE FROM camera
WHERE camera_id = ?;

-- name: CountCameras :one
SELECT COUNT(*)
FROM camera;

-- name: ListOnlineCameras :many
SELECT *
FROM camera
WHERE status = 'online'
ORDER BY camera_name;
