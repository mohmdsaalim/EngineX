package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngineGetOrCreateBook(t *testing.T) {
	e := NewEngine(nil)

	book1 := e.getOrCreateBook("INFY")
	book2 := e.getOrCreateBook("INFY")
	book3 := e.getOrCreateBook("RELIANCE")

	// Same symbol returns same book
	assert.Same(t, book1, book2)

	// Different symbol returns different book
	assert.NotSame(t, book1, book3)
	assert.Len(t, e.books, 2)
}

func TestParseSide(t *testing.T) {
	assert.Equal(t, Buy, parseSide("BUY"))
	assert.Equal(t, Sell, parseSide("SELL"))
	assert.Equal(t, Buy, parseSide("")) // default
}

func TestParseType(t *testing.T) {
	assert.Equal(t, Limit, parseType("LIMIT"))
	assert.Equal(t, Market, parseType("MARKET"))
	assert.Equal(t, Limit, parseType("")) // default
}