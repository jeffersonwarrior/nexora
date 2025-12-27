-- name: CreateCheckpoint :one
INSERT INTO checkpoints (
    id,
    session_id,
    timestamp,
    token_count,
    message_count,
    context_hash,
    state,
    compressed,
    created_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    strftime('%s', 'now')
) RETURNING *;

-- name: GetCheckpoint :one
SELECT *
FROM checkpoints
WHERE id = ? LIMIT 1;

-- name: GetLatestCheckpoint :one
SELECT *
FROM checkpoints
WHERE session_id = ?
ORDER BY timestamp DESC
LIMIT 1;

-- name: ListCheckpoints :many
SELECT *
FROM checkpoints
WHERE session_id = ?
ORDER BY timestamp DESC;

-- name: DeleteCheckpoint :exec
DELETE FROM checkpoints
WHERE id = ?;

-- name: DeleteOldCheckpoints :exec
DELETE FROM checkpoints
WHERE checkpoints.session_id = ?
AND checkpoints.id NOT IN (
    SELECT c.id
    FROM checkpoints AS c
    WHERE c.session_id = ?
    ORDER BY c.timestamp DESC
    LIMIT ?
);
