CREATE TABLE trades (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buy_order_id  UUID NOT NULL REFERENCES orders(id),
    sell_order_id UUID NOT NULL REFERENCES orders(id),
    buyer_id      UUID NOT NULL REFERENCES users(id),
    seller_id     UUID NOT NULL REFERENCES users(id),
    symbol        VARCHAR(20) NOT NULL,
    price         BIGINT NOT NULL,   -- scaled int
    quantity      BIGINT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trades_buyer_id    ON trades(buyer_id);
CREATE INDEX idx_trades_seller_id   ON trades(seller_id);
CREATE INDEX idx_trades_symbol      ON trades(symbol);
CREATE INDEX idx_trades_created_at  ON trades(created_at);