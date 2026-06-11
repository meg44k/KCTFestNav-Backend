 -- db/query.sql

-- ==========================================
-- Lives (ライブイベント関連)
-- ==========================================

-- name: GetLiveByID :one
SELECT * FROM lives WHERE id = ?;

-- name: GetAllLives :many
SELECT * FROM lives;

-- name: GetCurrentLives :many
-- statusの値は仮ですが、例えば1を「進行中」とした場合
SELECT * FROM lives WHERE status = 1;

-- name: CreateLive :exec
INSERT INTO lives (
name, detail, thumbnailURL, start_time, end_time, session_number, status
) VALUES (
?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateLive :exec
UPDATE lives
SET name = ?, detail = ?, thumbnailURL = ?, start_time = ?, end_time = ?, session_number = ?, status = ?
WHERE id = ?;

-- name: DeleteLive :exec
DELETE FROM lives WHERE id = ?;


-- ==========================================
-- Booths (ブース関連)
-- ==========================================

-- name: GetBoothByID :one
SELECT * FROM booths WHERE id = ?;

-- name: GetAllBooths :many
SELECT * FROM booths;

-- name: CreateBooth :execresult
INSERT INTO booths (
name, organizer, detail, congestion_status, x, y, z
) VALUES (
?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateBooth :exec
UPDATE booths
SET name = ?, organizer = ?, detail = ?, congestion_status = ?, x = ?, y = ?, z = ?
WHERE id = ?;

-- name: DeleteBooth :exec
DELETE FROM booths WHERE id = ?;


-- ==========================================
-- Users (ユーザー関連)
-- ==========================================

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: CreateUser :exec
-- usersテーブルはIDがUUID(VARCHAR)の指定だったため、挿入時にIDも受け取ります
INSERT INTO users (
id, name, assigned_booth_id, password, role
) VALUES (
?, ?, ?, ?, ?
);

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;