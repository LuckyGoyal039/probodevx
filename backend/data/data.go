package data

// inr balances

type User struct {
	Balance float32 `json:"balance"`
	Locked  float32 `json:"locked"`
}

var INR_BALANCES map[string]User = make(map[string]User)

// orderbook
type Order map[string]int
type PriceOptions struct {
	Total  int   `json:"total"`
	Orders Order `json:"orders"`
}
type OrderYesNo map[string]PriceOptions
type OrderSymbol struct {
	Yes OrderYesNo `json:"yes"`
	No  OrderYesNo `json:"no"`
}

var ORDERBOOK map[string]OrderSymbol = make(map[string]OrderSymbol)

// stock balances
type YesNo struct {
	Quantity int `json:"quantity"`
	Locked   int `json:"locked"`
}

type StockOption struct {
	Yes YesNo `json:"yes"`
	No  YesNo `json:"no"`
}
type UserStockBalance map[string]StockOption

var STOCK_BALANCES map[string]UserStockBalance = make(map[string]UserStockBalance)
