 -- db/query.sql

-- ==========================================
-- Stage (ステージイベント: セクション → ブロック → 出演者)
-- ==========================================

-- name: ListStageSections :many
SELECT * FROM stage_sections ORDER BY sort_order, id;

-- name: ListStageBlocks :many
SELECT * FROM stage_blocks ORDER BY start_time, id;

-- name: ListPerformers :many
SELECT * FROM performers ORDER BY block_id, perform_order, id;

-- name: ListPerformersInBlock :many
SELECT * FROM performers WHERE block_id = ? ORDER BY perform_order, id;

-- name: GetStageSection :one
SELECT * FROM stage_sections WHERE id = ?;

-- name: GetStageBlock :one
SELECT * FROM stage_blocks WHERE id = ?;

-- name: GetPerformer :one
SELECT * FROM performers WHERE id = ?;

-- name: CreateStageSection :execresult
INSERT INTO stage_sections (name, location, sort_order) VALUES (?, ?, ?);

-- name: UpdateStageSection :exec
UPDATE stage_sections SET name = ?, location = ?, sort_order = ? WHERE id = ?;

-- name: DeleteStageSection :exec
DELETE FROM stage_sections WHERE id = ?;

-- name: CreateStageBlock :execresult
INSERT INTO stage_blocks (section_id, start_time, end_time) VALUES (?, ?, ?);

-- name: UpdateStageBlock :exec
UPDATE stage_blocks SET start_time = ?, end_time = ? WHERE id = ?;

-- name: DeleteStageBlock :exec
DELETE FROM stage_blocks WHERE id = ?;

-- name: NextPerformOrder :one
SELECT CAST(COALESCE(MAX(perform_order), 0) + 1 AS SIGNED) AS next_order FROM performers WHERE block_id = ?;

-- name: CreatePerformer :execresult
INSERT INTO performers (block_id, name, detail, thumbnail_url, perform_order) VALUES (?, ?, ?, ?, ?);

-- name: UpdatePerformer :exec
UPDATE performers SET name = ?, detail = ?, thumbnail_url = ? WHERE id = ?;

-- name: SetPerformerOrder :exec
UPDATE performers SET perform_order = ? WHERE id = ?;

-- name: DeletePerformer :exec
DELETE FROM performers WHERE id = ?;

-- 同時に押されても出演者数 + 1(終了)を超えないよう、1 文で上限を取る
-- name: AdvanceBlock :exec
UPDATE stage_blocks
SET current_order = LEAST(current_order + 1,
  (SELECT COUNT(*) FROM performers WHERE performers.block_id = sqlc.arg(id)) + 1)
WHERE stage_blocks.id = sqlc.arg(id);

-- name: RewindBlock :exec
UPDATE stage_blocks SET current_order = GREATEST(current_order - 1, 0) WHERE id = ?;


-- ==========================================
-- Booths (ブース関連)
-- ==========================================

-- name: GetBoothByID :one
SELECT * FROM booths WHERE id = ?;

-- name: GetAllBooths :many
SELECT * FROM booths;

-- name: CreateBooth :exec
INSERT INTO booths (
name, organizer, detail, location, image_url, x, y, z, latitude, longitude
) VALUES (
?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateBooth :exec
UPDATE booths
SET name = ?, organizer = ?, detail = ?, location = ?, image_url = ?, x = ?, y = ?, z = ?, latitude = ?, longitude = ?
WHERE id = ?;

-- name: DeleteBooth :exec
DELETE FROM booths WHERE id = ?;


-- ==========================================
-- Users (ユーザー関連)
-- ==========================================

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByLoginID :one
SELECT * FROM users WHERE login_id = ?;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: CreateUser :exec
INSERT INTO users (
id, login_id, name, assigned_booth_id, password, role
) VALUES (
?, ?, ?, ?, ?, ?
);

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: UpdateUser :exec
UPDATE users
SET
    login_id = ?,
    name = ?,
    assigned_booth_id = ?,
    password = ?,
    role = ?
WHERE id = ?;