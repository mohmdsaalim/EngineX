CREATE TABLE balances (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) UNIQUE,
    available   BIGINT NOT NULL DEFAULT 0,   -- cash available to trade
    locked      BIGINT NOT NULL DEFAULT 0,   -- cash locked in open orders
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE positions (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id),
    symbol      VARCHAR(20) NOT NULL,
    quantity    BIGINT NOT NULL DEFAULT 0,   -- shares held
    locked_qty  BIGINT NOT NULL DEFAULT 0,  -- shares locked in open sell orders
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, symbol)
);

CREATE INDEX idx_balances_user_id   ON balances(user_id);
CREATE INDEX idx_positions_user_id  ON positions(user_id);
CREATE INDEX idx_positions_symbol   ON positions(symbol);