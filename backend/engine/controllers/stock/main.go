package stock

import (
	"context"
	"fmt"

	"github.com/probodevx/engine/global"
	"github.com/probodevx/engine/shared"
)

func GetStockBalances(ctx context.Context, event shared.EventModel) (interface{}, error) {
	userId := event.UserId
	if userId == "" {
		return global.StockManager.GetAllStockBalances(), nil
	}
	newData, exists := global.StockManager.GetStockBalances(userId)

	if !exists {
		return nil, fmt.Errorf("User not found")
	}
	return newData, nil
}

func CreateStock(ctx context.Context, event shared.EventModel) (interface{}, error) {

	stockSymbol := event.Data["stockSymbol"].(string)

	_, exists := global.OrderBookManager.GetOrderBook(stockSymbol)

	if exists {
		return nil, fmt.Errorf("Symbol %s already exists", stockSymbol)
	}

	global.OrderBookManager.AddOrderBookSymbol(stockSymbol)

	return map[string]interface{}{
		"message": fmt.Sprintf("Symbol %s created", stockSymbol),
	}, nil
}
