-- ENUS
CREATE TYPE order_side AS ENUM ('BUY', 'SELL'); -->  Prevents invalid value, only accpt the buy and sells
CREATE TYPE order_type AS ENUM ('LIMIT', 'MARKET'); 
CREATE TYPE order_status AS ENUM ('OPEN', 'PARTIAL', 'FILLED', 'CANCELLED');

CREATE TABLE orders (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id),
    symbol       VARCHAR(20) NOT NULL,
    side         order_side NOT NULL,
    type         order_type NOT NULL,
    price        BIGINT NOT NULL DEFAULT 0,  -- scaled int: ₹1500.50 = 150050
    quantity     BIGINT NOT NULL,
    filled_qty   BIGINT NOT NULL DEFAULT 0,
    status       order_status NOT NULL DEFAULT 'OPEN',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id  ON orders(user_id);
CREATE INDEX idx_orders_symbol   ON orders(symbol);
CREATE INDEX idx_orders_status   ON orders(status);