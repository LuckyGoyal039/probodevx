package orderbook

import (
	"context"
	"fmt"
	"strconv"

	"github.com/probodevx/engine/global"
	"github.com/probodevx/engine/shared"
)

func GetOrderbookSymbol(ctx context.Context, event shared.EventModel) (interface{}, error) {
	stockSymbol := event.Data["stockSymbol"].(string)
	if stockSymbol == "" {
		return global.OrderBookManager.GetAllOrderBook(), nil
	}
	newData, exists := global.OrderBookManager.GetOrderBook(stockSymbol)

	if !exists {
		return map[string]interface{}{"message": "stock symbol not found"}, nil
	}
	return newData, nil
}

func SellOrder(ctx context.Context, event shared.EventModel) (interface{}, error) {
	userId := event.UserId
	stockSymbol := event.Data["stockSymbol"].(string)
	quantityFloat := event.Data["quantity"].(float64)
	priceFloat := event.Data["price"].(float64)
	stockType := event.Data["stockType"].(string)
	quantity := int(quantityFloat)
	price := int(priceFloat)
	ok := checkValidStockBalance(userId, stockSymbol, stockType, quantity)
	if !ok {
		return nil, fmt.Errorf("Insufficient stock balance")
	}
	var reverseStockType string
	if stockType == "yes" {
		reverseStockType = "no"
	} else {
		reverseStockType = "yes"
	}
	reversePrice := 1000 - price
	priceList := GetValidSellPrices(stockSymbol, reverseStockType, reversePrice)
	remainingQuantity := quantity
	for _, currentPrice := range priceList {
		priceStr := strconv.FormatInt(int64(currentPrice), 10)
		priceData := global.OrderBookManager.GetPriceMap(stockSymbol, reverseStockType, priceStr)

		for sellerId, orderInfo := range priceData.Orders {
			if remainingQuantity <= 0 {
				break
			}
			availableQuantity := orderInfo.Quantity
			quantityToTake := min(remainingQuantity, availableQuantity)

			if quantityToTake > 0 {

				amount := quantityToTake * price
				global.UserManager.CreditBalance(userId, amount)
				global.UserManager.ReduceInrLock(sellerId, amount) // remove lock from the buyer (reverse order)

				if exists := global.StockManager.CheckUser(sellerId); !exists {
					global.StockManager.AddNewUser(sellerId)
					global.StockManager.AddStockBalancesSymbol(stockSymbol)
				}

				stockQty, _ := global.StockManager.GetQuantityStocks(sellerId, stockSymbol, stockType)
				stockQty += quantityToTake
				global.StockManager.SetStocksQuantity(sellerId, stockSymbol, stockType, stockQty)
				sellerStockQty, _ := global.StockManager.GetQuantityStocks(userId, stockSymbol, stockType)
				sellerStockQty -= quantityToTake
				global.StockManager.SetStocksQuantity(userId, stockSymbol, stockType, sellerStockQty)

				global.OrderBookManager.DecreaseUserQuantity(stockSymbol, reverseStockType, priceStr, sellerId, quantityToTake)
				global.OrderBookManager.DecreaseTotal(stockSymbol, reverseStockType, priceStr, quantityToTake)
				remainingQuantity -= quantityToTake
			}
		}
		priceData = global.OrderBookManager.GetPriceMap(stockSymbol, reverseStockType, priceStr)
		if priceData.Total <= 0 {
			global.OrderBookManager.RemovePrice(stockSymbol, stockType, priceStr)
		}

	}
	if remainingQuantity > 0 {
		PlaceSellOrder(stockSymbol, price, quantity, stockType, userId)
	}

	return map[string]interface{}{
		"message": fmt.Sprintf("Sell order placed for %v '%s' options at price %v.", quantity, stockType, price),
	}, nil
}

func BuyOrder(ctx context.Context, event shared.EventModel) (interface{}, error) {
	userId := event.UserId
	stockSymbol := event.Data["stockSymbol"].(string)
	quantityFloat := event.Data["quantity"].(float64)
	priceFloat := event.Data["price"].(float64)
	stockType := event.Data["stockType"].(string)
	quantity := int(quantityFloat)
	price := int(priceFloat)
	if ok := checkValidBalance(userId, price, quantity); !ok {
		return nil, fmt.Errorf("Insufficient INR balance")
	}
	priceList := GetValidPrices(stockSymbol, stockType, price)
	remainingQuantity := quantity
	for _, currentPrice := range priceList {
		priceStr := strconv.FormatInt(int64(currentPrice), 10)
		priceData := global.OrderBookManager.GetPriceMap(stockSymbol, stockType, priceStr)

		for sellerId, orderInfo := range priceData.Orders {
			if remainingQuantity <= 0 {
				break
			}
			availableQuantity := orderInfo.Quantity
			quantityToTake := min(remainingQuantity, availableQuantity)

			if quantityToTake > 0 {

				amount := quantityToTake * price
				global.UserManager.DebitBalance(userId, amount) //debit the qty of user
				if orderInfo.Reverse {
					lockAmt, _ := global.UserManager.GetUserLocked(sellerId)
					lockAmt -= quantityToTake * currentPrice
					global.UserManager.UpdateUserInrLock(sellerId, lockAmt)
					// global.UserManager.CreditBalance(sellerId, amount) //credit the quantity of seller
					global.StockManager.AddNewUser(sellerId)
					global.StockManager.AddStockBalancesSymbol(stockSymbol)
					var reverseStock string = "yes"
					if stockType == "yes" {
						reverseStock = "no"
					}
					qty, _ := global.StockManager.GetQuantityStocks(sellerId, stockSymbol, reverseStock)
					qty += quantityToTake
					global.StockManager.SetStocksQuantity(sellerId, stockSymbol, reverseStock, qty)
				} else {
					global.UserManager.CreditBalance(sellerId, amount) //credit the quantity of seller
					//remove lock of seller in stock_balances
					lockedQty, _ := global.StockManager.GetLockedStocks(sellerId, stockSymbol, stockType)
					lockedQty -= quantityToTake
					global.StockManager.SetStocksLock(sellerId, stockSymbol, stockType, lockedQty)
				}

				// Update buyer's stock balance
				AddStocksToBuyer(userId, stockSymbol, stockType, quantity)

				global.OrderBookManager.DecreaseUserQuantity(stockSymbol, stockType, priceStr, sellerId, quantityToTake)
				global.OrderBookManager.DecreaseTotal(stockSymbol, stockType, priceStr, quantityToTake)
				remainingQuantity -= quantityToTake
			}
		}
		priceData = global.OrderBookManager.GetPriceMap(stockSymbol, stockType, priceStr)
		if priceData.Total <= 0 {
			global.OrderBookManager.RemovePrice(stockSymbol, stockType, priceStr)
		}
	}
	if remainingQuantity > 0 {
		PlaceReverseBuyOrder(stockSymbol, price, quantity, stockType, userId)
	}
	if orderbookData, exists := global.OrderBookManager.GetOrderBook(stockSymbol); exists {
		err := PushInQueue(stockSymbol, orderbookData)
		if err != nil {
			panic(fmt.Sprintf("error: %s", err))
		}
	}
	return map[string]interface{}{"message": "Buy order placed and trade executed"}, nil
}

func CancelOrder(ctx context.Context, event shared.EventModel) (interface{}, error) {
	userId := event.UserId
	stockSymbol := event.Data["stockSymbol"].(string)
	quantityFloat := event.Data["quantity"].(float64)
	priceFloat := event.Data["price"].(float64)
	stockType := event.Data["stockType"].(string)
	quantity := int(quantityFloat)
	price := int(priceFloat)

	priceStr := strconv.FormatInt(int64(price), 10)
	priceData := global.OrderBookManager.GetPriceMap(stockSymbol, stockType, priceStr)

	global.OrderBookManager.DecreaseTotal(stockSymbol, stockType, priceStr, quantity)
	global.OrderBookManager.DecreaseUserQuantity(stockSymbol, stockType, priceStr, userId, quantity)

	priceData = global.OrderBookManager.GetPriceMap(stockSymbol, stockType, priceStr)
	if priceData.Total <= 0 {
		global.OrderBookManager.RemovePrice(stockSymbol, stockType, priceStr)
	}

	stockData, exists := global.StockManager.GetStockSymbol(userId, stockSymbol, stockType)
	if !exists {
		return map[string]interface{}{
			"message": "Invalid Input",
		}, nil
	}
	newLock := stockData.Locked - quantity
	global.StockManager.SetStocksLock(userId, stockSymbol, stockType, newLock)
	qty := stockData.Quantity + quantity
	global.StockManager.SetStocksQuantity(userId, stockSymbol, stockType, qty)

	return map[string]interface{}{
		"message": "Sell order canceled",
	}, nil
}
