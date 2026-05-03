package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mohmdsaalim/EngineX/internal/settlement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_NewExecutor(t *testing.T) {
	exec := settlement.NewExecutor(nil, nil)
	assert.NotNil(t, exec)
}

func TestExecutor_TradeMessage_Serialization(t *testing.T) {
	trade := settlement.TradeMessage{
		ID:          "trade-1",
		BuyOrderID:  "buy-o1",
		SellOrderID: "sell-o1",
		BuyerID:    "buyer-1",
		SellerID:   "seller-1",
		Symbol:     "INFY",
		Price:      150000,
		Quantity:   100,
		ExecutedAt: time.Now(),
	}

	payload, err := json.Marshal(trade)
	require.NoError(t, err)

	var result settlement.TradeMessage
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, trade.ID, result.ID)
	assert.Equal(t, trade.BuyOrderID, result.BuyOrderID)
	assert.Equal(t, trade.SellOrderID, result.SellOrderID)
	assert.Equal(t, trade.BuyerID, result.BuyerID)
	assert.Equal(t, trade.SellerID, result.SellerID)
	assert.Equal(t, trade.Symbol, result.Symbol)
	assert.Equal(t, trade.Price, result.Price)
	assert.Equal(t, trade.Quantity, result.Quantity)
}

func TestExecutor_TradeMessage_JSON(t *testing.T) {
	tests := []struct {
		name string
		trade settlement.TradeMessage
	}{
		{
			name: "basic trade",
			trade: settlement.TradeMessage{
				ID:          "basic-trade",
				BuyOrderID:  "buy-order",
				SellOrderID: "sell-order",
				BuyerID:    "buyer-user",
				SellerID:   "seller-user",
				Symbol:     "RELIANCE",
				Price:      250000,
				Quantity:   50,
			},
		},
		{
			name: "trade with high price",
			trade: settlement.TradeMessage{
				ID:          "high-price",
				BuyOrderID:  "buy-hp",
				SellOrderID: "sell-hp",
				BuyerID:    "buyer-hp",
				SellerID:   "seller-hp",
				Symbol:     "TCS",
				Price:      500000,
				Quantity:   10,
			},
		},
		{
			name: "trade large quantity",
			trade: settlement.TradeMessage{
				ID:          "large-qty",
				BuyOrderID:  "buy-lq",
				SellOrderID: "sell-lq",
				BuyerID:    "buyer-lq",
				SellerID:   "seller-lq",
				Symbol:     "NIFTY",
				Price:      22000,
				Quantity:   1000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(tt.trade)
			require.NoError(t, err)

			var result settlement.TradeMessage
			err = json.Unmarshal(payload, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.trade.ID, result.ID)
			assert.Equal(t, tt.trade.Symbol, result.Symbol)
			assert.Equal(t, tt.trade.Price, result.Price)
			assert.Equal(t, tt.trade.Quantity, result.Quantity)
		})
	}
}

func TestExecutor_TradeValueCalculation(t *testing.T) {
	tests := []struct {
		name     string
		price   int64
		qty     int64
		expected int64
	}{
		{
			name:     "small trade",
			price:    150000,
			qty:      100,
			expected: 15000000,
		},
		{
			name:     "medium trade",
			price:    250000,
			qty:      50,
			expected: 12500000,
		},
		{
			name:     "large trade",
			price:    50000,
			qty:      1000,
			expected: 50000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trade := settlement.TradeMessage{
				ID:        "test",
				Price:    tt.price,
				Quantity: tt.qty,
			}
			tradeValue := trade.Price * trade.Quantity
			assert.Equal(t, tt.expected, tradeValue)
		})
	}
}

func TestExecutor_ProcessTrade_InvalidJSON(t *testing.T) {
	exec := settlement.NewExecutor(nil, nil)

	invalidJSON := []byte(`{invalid`)
	err := exec.ProcessTrade(context.Background(), invalidJSON)
	assert.Error(t, err)
}

func TestExecutor_ProcessTrade_EmptyJSON(t *testing.T) {
	exec := settlement.NewExecutor(nil, nil)

	emptyJSON := []byte(`{}`)
	err := exec.ProcessTrade(context.Background(), emptyJSON)
	assert.Error(t, err)
}

func TestExecutor_TradeMessage_AllFields(t *testing.T) {
	trade := settlement.TradeMessage{
		ID:          "full-trade",
		BuyOrderID:  "buy-full",
		SellOrderID: "sell-full",
		BuyerID:    "buyer-full",
		SellerID:   "seller-full",
		Symbol:     "ADANI",
		Price:      150050,
		Quantity:   25,
		ExecutedAt: time.Now(),
	}

	payload, err := json.Marshal(trade)
	require.NoError(t, err)

	var result settlement.TradeMessage
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, "full-trade", result.ID)
	assert.Equal(t, "buy-full", result.BuyOrderID)
	assert.Equal(t, "sell-full", result.SellOrderID)
	assert.Equal(t, "buyer-full", result.BuyerID)
	assert.Equal(t, "seller-full", result.SellerID)
	assert.Equal(t, "ADANI", result.Symbol)
	assert.Equal(t, int64(150050), result.Price)
	assert.Equal(t, int64(25), result.Quantity)
}

func TestExecutor_MultipleTrades_Aggregation(t *testing.T) {
	trades := []settlement.TradeMessage{
		{ID: "t1", Price: 100, Quantity: 10},
		{ID: "t2", Price: 200, Quantity: 20},
		{ID: "t3", Price: 150, Quantity: 15},
	}

	var totalValue int64
	var totalQty int64

	for _, trade := range trades {
		totalValue += trade.Price * trade.Quantity
		totalQty += trade.Quantity
	}

	assert.Equal(t, int64(45), totalQty)
	assert.Equal(t, int64(7250), totalValue)
}