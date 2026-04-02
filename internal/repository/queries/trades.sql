-- name: CreateTrade :one
INSERT INTO trades (buy_order_id, sell_order_id, buyer_id, seller_id, symbol, price, quantity)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTradesBySymbol :many
SELECT * FROM trades
WHERE symbol = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetTradesByUserID :many
SELECT * FROM trades
WHERE buyer_id = $1 OR seller_id = $1
ORDER BY created_at DESC;
