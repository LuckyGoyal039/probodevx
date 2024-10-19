package data

type User struct {
	Balance float32 `json:"balance"`
	Locked  float32 `json:"locked"`
}

var INR_BALANCES map[string]User = make(map[string]User)

var ORDERBOOK map[string]interface{} = make(map[string]interface{})

var STOCK_BALANCES map[string]interface{}
