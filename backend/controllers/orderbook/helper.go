package orderbook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	redis "github.com/probodevx/config"
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
	} else if stockType == "no" {
		stockTypeData = orderData.No
	} else {
		return "none"
	}

	var availableQuantity int = 0
	hasValidPrice := false

	for priceStr, priceData := range stockTypeData {
		fmt.Printf("Original priceStr before ParseInt: %v\n", priceStr)
		currentPrice64, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing priceStr: %v\n", err)
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

	orderData, _ := global.OrderBookManager.GetOrderBook(stockSymbol)

	var stockTypeData data.OrderYesNo
	if stockType == "yes" {
		stockTypeData = orderData.Yes
	} else if stockType == "no" {
		stockTypeData = orderData.No
	}

	remainingQuantity := quantity
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

			availableQuantity := orderInfo.Quantity
			quantityToTake := min(remainingQuantity, availableQuantity)

			if quantityToTake > 0 {
				//remove lock of seller in stock_balances
				lockedQty, _ := global.StockManager.GetLockedStocks(sellerId, stockSymbol, stockType)
				lockedQty -= quantityToTake
				global.StockManager.SetStocksLock(sellerId, stockSymbol, stockType, lockedQty)

				// Update buyer's stock balance
				// check it can be optimized
				global.StockManager.AddNewUser(userId)
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
				global.StockManager.UpdateStockBalanceSymbol(userId, stockSymbol, buyerStockOption)

				priceData.Total -= quantityToTake
				orderInfo.Quantity -= quantityToTake
				if orderInfo.Quantity == 0 {
					delete(priceData.Orders, sellerId)
				}

				remainingQuantity -= quantityToTake

				//update seller inr_balance
				value, _ := global.UserManager.GetUserBalance(sellerId)
				totalSellerBalance := value + quantityToTake*currentPrice
				global.UserManager.UpdateUserInrBalance(sellerId, totalSellerBalance)
			}
		}

		// Remove price level if no orders remain
		if len(priceData.Orders) == 0 {
			delete(stockTypeData, priceStr)
		}
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
	totalAmt := quantity * price
	global.UserManager.DebitBalance(userId, totalAmt)
	lockAmt, _ := global.UserManager.GetUserLocked(userId)
	lockAmt += totalAmt
	global.UserManager.UpdateUserInrLock(userId, lockAmt)
	return nil
}

func checkAndLockBalance(userId string, price int, quantity int) (bool, error) {
	user, exists := global.UserManager.GetUser(userId)
	if !exists {
		return false, fmt.Errorf("user not found")
	}

	totalCost := price * quantity

	if user.Balance < totalCost {
		return false, fmt.Errorf("Insufficient INR balance")
	}
	leftBalance := user.Balance - totalCost
	lockedAmount := user.Locked + totalCost

	global.UserManager.UpdateUserInrBalance(userId, leftBalance)
	global.UserManager.UpdateUserInrLock(userId, lockedAmount)

	return true, nil
}
func checkValidBalance(userId string, price int, quantity int) bool {
	balance, exists := global.UserManager.GetUserBalance(userId)
	if !exists {
		return false
	}
	totalCost := price * quantity
	return balance >= totalCost
}
func UnLockBalance(userId string, quantity, price int) (bool, error) {
	user, exists := global.UserManager.GetUser(userId)
	if !exists {
		return false, fmt.Errorf("user not found")
	}

	lockedAmount := max(0, user.Locked-quantity*price)
	global.UserManager.UpdateUserInrLock(userId, lockedAmount)
	return true, nil
}

func PushInQueue(stockSymbol string, orderbookData data.OrderSymbol) error {

	if err := redis.CheckRedisConnection(); err != nil {
		return fmt.Errorf("redis connection error: %v", err)
	}
	redisClient := redis.GetRedisClient()

	ctx := context.TODO()

	jsonData, err := json.Marshal(orderbookData)
	if err != nil {
		return fmt.Errorf("error marshaling orderbook data %s", err)
	}
	queryKey := fmt.Sprintf("orderbook:%s", stockSymbol)

	if redis.Redis == nil {
		return fmt.Errorf("Redis client is nil. Please ensure Redis is properly initialized")
	}

	if _, err := redisClient.LPush(ctx, queryKey, jsonData).Result(); err != nil {
		return fmt.Errorf("error pushing in redis: %s", err)
	}

	return nil
}

func CheckAndLockStock(userId string, stockSymbol string, stockType string, quantity int) bool {
	exist := global.StockManager.CheckUser(userId)
	if !exist {
		return false
	}
	availableQuantity, err := global.StockManager.GetQuantityStocks(userId, stockSymbol, stockType)
	if err != nil {
		return false
	}
	if availableQuantity < quantity {
		return false
	}
	//reduce qty also

	updatedQty := availableQuantity - quantity
	if err := global.StockManager.SetStocksQuantity(userId, stockSymbol, stockType, updatedQty); err != nil {
		return false
	}
	if err := global.StockManager.SetStocksLock(userId, stockSymbol, stockType, quantity); err != nil {
		return false
	}
	return true
}

func CheckBuyer(stockSymbol string, stockType string, price int, quantity int) string {
	orderData, exists := global.OrderBookManager.GetOrderBook(stockSymbol)
	if !exists {
		return "none"
	}

	var stockTypeData data.OrderYesNo
	if stockType == "yes" {
		stockTypeData = orderData.No
	} else if stockType == "no" {
		stockTypeData = orderData.Yes
	} else {
		return "none"
	}

	var availableQuantity int = 0
	hasValidPrice := false

	for priceStr, priceData := range stockTypeData {
		fmt.Printf("Original priceStr before ParseInt: %v\n", priceStr)
		currentPrice64, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing priceStr: %v\n", err)
			continue
		}
		currentPrice := int(currentPrice64)

		if currentPrice >= price {
			for _, orders := range priceData.Orders {
				if orders.Reverse {
					hasValidPrice = true
					total := orders.Quantity
					if total > 0 {
						availableQuantity += total
					}
				}
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

func FullFillSellOrder(stockSymbol string, price, quantity int, stockType, userId string) error {
	orderData, _ := global.OrderBookManager.GetOrderBook(stockSymbol)
	reverseStockType := map[string]string{"yes": "no", "no": "yes"}[stockType]

	stockTypeData := orderData.No
	if reverseStockType == "yes" {
		stockTypeData = orderData.Yes
	}

	prices := make([]int, 0, len(stockTypeData))
	for priceStr, priceData := range stockTypeData {
		currPrice, err := strconv.Atoi(priceStr)
		if err != nil || currPrice > price {
			continue
		}
		for _, order := range priceData.Orders {
			if order.Reverse {
				prices = append(prices, currPrice)
				break
			}
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(prices)))

	remainingQuantity := quantity
	for _, currentPrice := range prices {
		if remainingQuantity <= 0 {
			break
		}

		priceStr := strconv.Itoa(currentPrice)
		priceData := stockTypeData[priceStr]

		for sellerId, orderInfo := range priceData.Orders {
			if remainingQuantity <= 0 || !orderInfo.Reverse {
				break
			}

			quantityToTake := min(remainingQuantity, orderInfo.Quantity)
			remainingQuantity -= quantityToTake
			restQty := max(0, orderInfo.Quantity-quantityToTake)
			global.StockManager.SetStocksQuantity(sellerId, stockSymbol, reverseStockType, restQty)

			currentLock, _ := global.UserManager.GetUserLocked(sellerId)
			global.UserManager.UpdateUserInrLock(sellerId, currentLock-restQty*currentPrice)

			if restQty == 0 {
				global.OrderBookManager.RemoveUserFromOrder(sellerId, stockSymbol, stockType, currentPrice)
			} else {
				global.OrderBookManager.UpdateStockQtyFromOrder(sellerId, stockSymbol, stockType, currentPrice, restQty)
			}
		}

		if len(priceData.Orders) == 0 {
			delete(stockTypeData, priceStr)
		}
	}
	return nil
}

func PlacePartialSellOrder(stockSymbol string, price int, quantity int, stockType string, userId string) {
	// Calculate fulfillable quantity
	fulfillableQuantity := calculateFulfillableQuantity(stockSymbol, price, quantity, stockType)
	remainingQuantity := quantity - fulfillableQuantity

	// Fulfill as much of the order as possible
	if fulfillableQuantity > 0 {
		err := FullFillSellOrder(stockSymbol, price, fulfillableQuantity, stockType, userId)
		if err != nil {
			// Handle error if needed
			return
		}
	}

	// Place remaining quantity as a new sell order
	if remainingQuantity > 0 {
		PlaceSellOrder(stockSymbol, price, remainingQuantity, stockType, userId)
	}
}

func calculateFulfillableQuantity(stockSymbol string, price int, quantity int, stockType string) int {
	orderData, _ := global.OrderBookManager.GetOrderBook(stockSymbol)
	reverseStockType := map[string]string{"yes": "no", "no": "yes"}[stockType]

	stockTypeData := orderData.No
	if reverseStockType == "yes" {
		stockTypeData = orderData.Yes
	}

	prices := make([]int, 0, len(stockTypeData))
	for priceStr, priceData := range stockTypeData {
		currPrice, err := strconv.Atoi(priceStr)
		if err != nil || currPrice > price {
			continue
		}
		for _, order := range priceData.Orders {
			if order.Reverse {
				prices = append(prices, currPrice)
				break
			}
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(prices)))

	remainingQuantity := quantity
	fulfillableQuantity := 0

	for _, currentPrice := range prices {
		if remainingQuantity <= 0 {
			break
		}

		priceStr := strconv.Itoa(currentPrice)
		priceData := stockTypeData[priceStr]

		for _, orderInfo := range priceData.Orders {
			if remainingQuantity <= 0 || !orderInfo.Reverse {
				break
			}

			quantityToTake := min(remainingQuantity, orderInfo.Quantity)
			remainingQuantity -= quantityToTake
			fulfillableQuantity += quantityToTake
		}
	}
	return fulfillableQuantity
}

func PlaceSellOrder(stockSymbol string, price, quantity int, stockType, userId string) {
	if _, exist := global.OrderBookManager.GetOrderBook(stockSymbol); !exist {
		global.OrderBookManager.AddOrderBookSymbol(stockSymbol)
	}
	global.OrderBookManager.AddOrderbookPrice(stockSymbol, stockType, price)
	global.OrderBookManager.UpdateSellOrder(userId, stockSymbol, stockType, price, quantity)
	locked, _ := global.StockManager.GetLockedStocks(userId, stockSymbol, stockType)
	locked += quantity
	global.StockManager.SetStocksLock(userId, stockSymbol, stockType, locked)
	balance, _ := global.StockManager.GetQuantityStocks(userId, stockSymbol, stockType)

	balance -= quantity
	global.StockManager.SetStocksQuantity(userId, stockSymbol, stockType, balance)
}

func GetValidPrices(stockSymbol string, stockType string, price int) []int {
	orderData, _ := global.OrderBookManager.GetOrderBook(stockSymbol)

	var stockTypeData data.OrderYesNo
	if stockType == "yes" {
		stockTypeData = orderData.Yes
	} else if stockType == "no" {
		stockTypeData = orderData.No
	}
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
	return prices
}

func AddStocksToBuyer(userId, stockSymbol, stockType string, quantity int) {
	global.StockManager.AddNewUser(userId)
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
		buyerStockOption.Yes.Quantity += quantity
	} else {
		buyerStockOption.No.Quantity += quantity
	}
	global.StockManager.UpdateStockBalanceSymbol(userId, stockSymbol, buyerStockOption)
}

func checkValidStockBalance(userId, stockSymbol, stockType string, quantityToSell int) bool {
	UserQty, err := global.StockManager.GetQuantityStocks(userId, stockSymbol, stockType)
	if err != nil {
		return false
	}
	return UserQty >= quantityToSell
}

func GetValidSellPrices(stockSymbol string, stockType string, price int) []int {
	orderData, _ := global.OrderBookManager.GetOrderBook(stockSymbol)

	var stockTypeData data.OrderYesNo
	if stockType == "yes" {
		stockTypeData = orderData.Yes
	} else if stockType == "no" {
		stockTypeData = orderData.No
	}
	var prices []int
	for priceStr, priceData := range stockTypeData {
		currPrice64, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			continue
		}
		for _, orderInfo := range priceData.Orders {
			if orderInfo.Reverse {
				currPrice := int(currPrice64)
				if currPrice >= price {
					prices = append(prices, currPrice)
				}
			}
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(prices)))
	return prices
}
