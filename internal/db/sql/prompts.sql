-- name: CreatePrompt :one
INSERT INTO prompt_library (
    id,
    category,
    subcategory,
    title,
    description,
    content,
    content_hash,
    tags,
    author
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPrompt :one
SELECT * FROM prompt_library WHERE id = ? LIMIT 1;

-- name: ListPrompts :many
SELECT * FROM prompt_library
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListPromptsByCategory :many
SELECT * FROM prompt_library
WHERE category = ?
ORDER BY rating DESC, usage_count DESC
LIMIT ?;

-- name: SearchPrompts :many
SELECT p.* FROM prompt_library p
WHERE p.rowid IN (
    SELECT f.rowid FROM prompt_library_fts f WHERE f.title MATCH ? OR f.description MATCH ? OR f.content MATCH ? OR f.tags MATCH ?
)
LIMIT ?;

-- name: UpdatePrompt :exec
UPDATE prompt_library
SET title = ?,
    description = ?,
    content = ?,
    content_hash = ?,
    tags = ?,
    updated_at = strftime('%s', 'now')
WHERE id = ?;

-- name: DeletePrompt :exec
DELETE FROM prompt_library WHERE id = ?;

-- name: IncrementUsage :exec
UPDATE prompt_library
SET usage_count = usage_count + 1,
    last_used_at = strftime('%s', 'now')
WHERE id = ?;

-- name: UpdateRating :exec
UPDATE prompt_library
SET rating = sqlc.arg(rating),
    votes = votes + 1,
    updated_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id);

-- name: ListTopPrompts :many
SELECT * FROM prompt_library
ORDER BY rating DESC, usage_count DESC
LIMIT ?;

-- name: GetPromptsByTag :many
SELECT * FROM prompt_library
WHERE tags LIKE '%' || sqlc.arg(tag) || '%'
ORDER BY rating DESC
LIMIT sqlc.arg(limit);
