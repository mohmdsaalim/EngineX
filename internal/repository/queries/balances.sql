-- name: GetBalanceByUserID :one
SELECT * FROM balances WHERE user_id = $1;

-- name: CreateBalance :one
INSERT INTO balances (user_id, available, locked)
VALUES ($1, $2, 0)
RETURNING *;


-- name: DebitBalance :one
UPDATE balances
SET available  = available - @debit_amount,
    locked     = locked + @debit_amount,
    updated_at = NOW()
WHERE user_id = @user_id
  AND available >= @debit_amount
RETURNING *;

-- name: CreditBalance :one
UPDATE balances
SET available  = available + @credit_amount,
    locked     = locked - @locked_release,
    updated_at = NOW()
WHERE user_id = @user_id
RETURNING *;