-- name: CreateOrder :one
INSERT INTO orders (user_id, symbol, side, type, price, quantity, status)
VALUES ($1, $2, $3, $4, $5, $6, 'OPEN')
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrdersByUserID :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2, filled_qty = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetOpenOrdersBySymbol :many
SELECT * FROM orders
WHERE symbol = $1 AND status IN ('OPEN', 'PARTIAL')
ORDER BY created_at ASC;