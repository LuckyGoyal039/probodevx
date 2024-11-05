package mint

import (
	"context"
	"fmt"

	"github.com/probodevx/engine/global"
	"github.com/probodevx/engine/shared"
)

func MintStock(ctx context.Context, event shared.EventModel) (interface{}, error) {

	userId := event.UserId
	stockSymbol := event.Data["stockSymbol"].(string)
	quantityFloat := event.Data["quantity"].(float64)
	priceFloat := event.Data["price"].(float64)
	price := int64(priceFloat)
	quantity := int64(quantityFloat)

	if err := updateBalance(int(price), int(quantity), userId); err != nil {
		return nil, fmt.Errorf("insufficient balance")
	}
	stockData, exist := global.StockManager.GetStockBalances(userId)
	if !exist {
		global.StockManager.AddNewUser(userId)
	}
	global.StockManager.AddStockBalancesSymbol(stockSymbol)
	data := stockData[stockSymbol]
	data.No.Quantity = int(quantity)
	data.Yes.Quantity = int(quantity)
	global.StockManager.UpdateStockBalanceSymbol(userId, stockSymbol, data)

	global.OrderBookManager.AddOrderBookSymbol(stockSymbol)

	currentBalance, exist := global.UserManager.GetUserBalance(userId)
	if !exist {
		return nil, fmt.Errorf("user not found")
	}
	return map[string]interface{}{
		"message": fmt.Sprintf("Minted %v 'yes' and 'no' tokens for user %s, remaining balance is %v", quantity, userId, currentBalance),
	}, nil
}
