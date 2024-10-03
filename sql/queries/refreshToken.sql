-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES(
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET 
    updated_at = NOW(),  -- Set RevokedAt to the current time
    revoked_at = NOW()   -- Set UpdatedAt to the current time
WHERE 
    token = $1; 

-- name: DeleteAllTokens :exec
DELETE FROM refresh_tokens;