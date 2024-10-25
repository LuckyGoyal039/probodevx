package orderbook

import (
	"errors"
	"sort"
	"strconv"

	"github.com/probodevx/data"
	"github.com/probodevx/global"
)

func CheckCanPlaceOrder(stockSymbol string, price int, quantity int, stockType string) string {

	orderData, exists := global.OrderBookManager.GetOrderBook(stockSymbol)
	if !exists {
		return "none"
	}

	var stockTypeData data.OrderYesNo
	if stockType == "yes" {
		stockTypeData = orderData.Yes
	} else if stockSymbol == "no" {
		stockTypeData = orderData.No
	} else {
		return "none"
	}

	var availableQuantity int = 0
	hasValidPrice := false

	for priceStr, priceData := range stockTypeData {

		currentPrice64, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			continue
		}
		currentPrice := int(currentPrice64)

		if currentPrice <= price {
			hasValidPrice = true
			total := priceData.Total
			if total > 0 {
				availableQuantity += total
			}
		}
	}

	if !hasValidPrice {
		return "none"
	}

	if availableQuantity >= quantity {
		return "fullfill"
	}

	if availableQuantity > 0 {
		return "partial"
	}

	return "none"
}

func PlaceFullFillOrder(stockSymbol string, price int, quantity int, stockType string, userId string) error {

	userBalance, exists := global.UserManager.GetUser(userId)
	if !exists {
		return errors.New("user not found in balance sheet")
	}

	totalCost := price * quantity
	availableBalance := userBalance.Balance
	if availableBalance < totalCost {
		return errors.New("insufficient balance")
	}

	orderData, exists := global.OrderBookManager.GetOrderBook(stockSymbol)
	if !exists {
		return errors.New("stock symbol not found")
	}

	var stockTypeData data.OrderYesNo
	if stockType == "yes" {
		stockTypeData = orderData.Yes
	} else if stockSymbol == "no" {
		stockTypeData = orderData.No
	}

	remainingQuantity := quantity
	costIncurred := 0

	var prices []int
	for priceStr := range stockTypeData {
		currPrice64, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			continue
		}
		currPrice := int(currPrice64)
		if currPrice <= price {
			prices = append(prices, currPrice)
		}
	}
	sort.Ints(prices)

	for _, currentPrice := range prices {
		if currentPrice > price {
			continue
		}

		priceStr := strconv.FormatInt(int64(currentPrice), 10)
		priceData := stockTypeData[priceStr]

		for sellerId, orderInfo := range priceData.Orders {
			if remainingQuantity <= 0 {
				break
			}

			// Skip if order is reversed
			// if orderInfo.reverse {
			// 	continue
			// }

			availableQuantity := orderInfo.Quantity
			quantityToTake := min(remainingQuantity, availableQuantity)

			if quantityToTake > 0 {
				// Update seller's stock balance
				sellerBalance, exists := global.StockManager.GetStockBalances(sellerId)
				if !exists {
					sellerBalance = make(data.UserStockBalance)
				}
				stockOption, exists := sellerBalance[stockSymbol]
				if !exists {
					stockOption = data.StockOption{
						Yes: data.YesNo{Quantity: 0, Locked: 0},
						No:  data.YesNo{Quantity: 0, Locked: 0},
					}
				}
				if stockType == "yes" {
					stockOption.Yes.Locked -= quantityToTake
				} else {
					stockOption.No.Locked -= quantityToTake
				}
				// sellerBalance[stockSymbol] = stockOption
				global.StockManager.UpdateStockBalanceSymbol(sellerId, stockSymbol, stockOption)
				// data.STOCK_BALANCES[sellerId] = sellerBalance

				// Update buyer's stock balance
				buyerBalance, exists := global.StockManager.GetStockBalances(userId)
				if !exists {
					buyerBalance = make(data.UserStockBalance)
				}

				buyerStockOption, exists := buyerBalance[stockSymbol]
				if !exists {
					buyerStockOption = data.StockOption{
						Yes: data.YesNo{Quantity: 0, Locked: 0},
						No:  data.YesNo{Quantity: 0, Locked: 0},
					}
				}

				if stockType == "yes" {
					buyerStockOption.Yes.Quantity += quantityToTake
				} else {
					buyerStockOption.No.Quantity += quantityToTake
				}
				// buyerBalance[stockSymbol] = buyerStockOption
				global.StockManager.UpdateStockBalanceSymbol(userId, stockSymbol, buyerStockOption)

				priceData.Total -= quantityToTake
				orderInfo.Quantity -= quantityToTake
				if orderInfo.Quantity == 0 {
					delete(priceData.Orders, sellerId)
				}

				remainingQuantity -= quantityToTake
				costIncurred += quantityToTake * currentPrice

				buyerInrBalance, exists := global.UserManager.GetUser(userId)
				sellerInrBalance, exists := global.UserManager.GetUser(sellerId)
				if !exists {
					panic("user not found")
				}

				// Update balances
				totalBuyerBalance := buyerInrBalance.Balance - costIncurred
				totalSellerBalance := sellerInrBalance.Balance + costIncurred

				// Save back to map
				global.UserManager.UpdateUserInrBalance(userId, totalBuyerBalance)
				global.UserManager.UpdateUserInrBalance(sellerId, totalSellerBalance)
			}
		}

		// Remove price level if no orders remain
		if len(priceData.Orders) == 0 {
			delete(stockTypeData, priceStr)
		}
	}

	if remainingQuantity > 0 {
		return errors.New("could not fulfill entire order")
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetFullFillableQuantity(stockSymbol string, price int, quantity int, stockType string) int {
	orderData, exists := global.OrderBookManager.GetOrderBook(stockSymbol)
	if !exists {
		return 0
	}
	var stockTypeData data.OrderYesNo
	if stockType == "yes" {
		stockTypeData = orderData.Yes
	} else if stockType == "no" {
		stockTypeData = orderData.No
	} else {
		return 0
	}

	availableQuantity := 0

	for priceStr, priceData := range stockTypeData {
		currentPrice64, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			continue
		}
		currentPrice := int(currentPrice64)

		if currentPrice <= price {
			total := priceData.Total
			if total > 0 {
				availableQuantity += total
				if availableQuantity >= quantity {
					return quantity
				}
			}
		}
	}
	return availableQuantity
}

func PlacePartialOrder(stockSymbol string, price int, quantity int, stockType string, userId string) {

	fullFillQty := GetFullFillableQuantity(stockSymbol, price, quantity, stockType)

	remainingQuantity := quantity - fullFillQty

	if fullFillQty > 0 {
		PlaceFullFillOrder(stockSymbol, price, quantity, stockType, userId)
	}

	if remainingQuantity > 0 {
		PlaceReverseBuyOrder(stockSymbol, price, quantity, stockType, userId)
	}

}

func PlaceReverseBuyOrder(stockSymbol string, price int, quantity int, stockType string, userId string) error {

	var reverseStockType string
	if stockType == "yes" {
		reverseStockType = "no"
	} else if stockType == "no" {
		reverseStockType = "yes"
	} else {
		return errors.New("invalid stock type")
	}

	reversePrice := 1000 - price
	if reversePrice < 0 {
		return errors.New("invalid price for reverse order")
	}
	global.OrderBookManager.CreateOrderbookPrice(stockSymbol, reverseStockType, reversePrice, quantity, userId, true)
	return nil
}
