-- name: CreateResource :one
INSERT INTO resources (name, description)
VALUES ($1, $2)
RETURNING id, name, description, created_at, updated_at;

-- name: GetResource :one
SELECT id, name, description, created_at, updated_at
FROM resources
WHERE id = $1;

-- name: ListResources :many
SELECT id, name, description, created_at, updated_at
FROM resources
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateResource :one
UPDATE resources
SET name = $1, description = $2, updated_at = now()
WHERE id = $3
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteResource :exec
DELETE FROM resources
WHERE id = $1;
