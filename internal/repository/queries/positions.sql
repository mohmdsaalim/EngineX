-- name: GetPositionByUserAndSymbol :one
SELECT * FROM positions
WHERE user_id = $1 AND symbol = $2;

-- name: UpsertPosition :one
INSERT INTO positions (user_id, symbol, quantity, locked_qty)
VALUES ($1, $2, $3, 0)
ON CONFLICT (user_id, symbol)
DO UPDATE SET
    quantity   = positions.quantity + EXCLUDED.quantity,
    updated_at = NOW()
RETURNING *;