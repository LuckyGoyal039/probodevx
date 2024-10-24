package orderbook

import (
	"errors"
	"sort"
	"strconv"

	"github.com/probodevx/data"
	"github.com/probodevx/global"
)

func CreateSymbolOrderbook(stockSymbol string) {
	availableSymbol, exists := data.ORDERBOOK[stockSymbol]
	if !exists {
		availableSymbol = data.OrderSymbol{
			Yes: make(data.OrderYesNo),
			No:  make(data.OrderYesNo),
		}
		data.ORDERBOOK[stockSymbol] = availableSymbol
	}
}

func CheckCanPlaceOrder(stockSymbol string, price float64, quantity int, stockType string) string {

	orderData, exists := data.ORDERBOOK[stockSymbol]
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

		currentPrice, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			continue
		}

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

func PlaceFullfillOrder(stockSymbol string, price float64, quantity int, stockType string, userId string) error {

	userBalance, exists := global.UserManager.INR_BALANCES[userId]
	if !exists {
		return errors.New("user not found in balance sheet")
	}

	totalCost := price * float64(quantity)
	availableBalance := float64(userBalance.Balance)
	if availableBalance < totalCost {
		return errors.New("insufficient balance")
	}

	orderData, exists := data.ORDERBOOK[stockSymbol]
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
	costIncurred := 0.0

	var prices []float64
	for priceStr := range stockTypeData {
		currPrice, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			continue
		}
		if currPrice <= price {
			prices = append(prices, currPrice)
		}
	}
	sort.Float64s(prices)

	for _, currentPrice := range prices {
		if currentPrice > price {
			continue
		}

		priceStr := strconv.FormatFloat(currentPrice, 'f', -1, 64)
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
				sellerBalance, exists := data.STOCK_BALANCES[sellerId]
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
				sellerBalance[stockSymbol] = stockOption
				data.STOCK_BALANCES[sellerId] = sellerBalance

				// Update buyer's stock balance
				buyerBalance, exists := data.STOCK_BALANCES[userId]
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
				buyerBalance[stockSymbol] = buyerStockOption
				data.STOCK_BALANCES[userId] = buyerBalance

				priceData.Total -= quantityToTake
				orderInfo.Quantity -= quantityToTake
				if orderInfo.Quantity == 0 {
					delete(priceData.Orders, sellerId)
				}

				remainingQuantity -= quantityToTake
				costIncurred += float64(quantityToTake) * currentPrice

				buyerInrBalance := global.UserManager.INR_BALANCES[userId]
				sellerInrBalance := global.UserManager.INR_BALANCES[sellerId]

				// Update balances
				buyerInrBalance.Balance -= float32(costIncurred)
				sellerInrBalance.Balance += float32(costIncurred)

				// Save back to map
				global.UserManager.INR_BALANCES[userId] = buyerInrBalance
				global.UserManager.INR_BALANCES[sellerId] = sellerInrBalance
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
func PlacePartialOrder() {

}
func PlaceReverseBuyOrder() {

}
