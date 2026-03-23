-- name: CreateChat :one
INSERT INTO chats (user_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: GetChatByID :one
SELECT * FROM chats
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: GetAllChatsByUserID :many
SELECT * FROM chats
WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: UpdateChatTitle :exec
UPDATE chats
SET title = $1, updated_at = NOW()
WHERE id = $2 AND user_id = $3;

-- name: DeleteChat :exec
DELETE FROM chats
WHERE id = $1 AND user_id = $2;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (chat_id, role, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetChatMessages :many
SELECT * FROM chat_messages
WHERE chat_id = $1
ORDER BY created_at ASC;
