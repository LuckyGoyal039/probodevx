// global/managers.go
package global

import "github.com/probodevx/engine/data"

var (
	UserManager      = data.NewUserManager()
	OrderBookManager = data.NewOrderBookManager()
	StockManager     = data.NewStockManager()
)
