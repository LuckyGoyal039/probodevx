package orderbook

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/controllers/common"
	"github.com/probodevx/data"
	"github.com/probodevx/global"
)

func GetOrderBook(c *fiber.Ctx) error {
	stockSymbol := c.Params("stockSymbol")
	if stockSymbol == "" {
		return c.JSON(data.ORDERBOOK)
	}
	newData, exists := data.ORDERBOOK[stockSymbol]

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("stock symbol not found")
	}
	return c.JSON(newData)
}

func SellOrder(c *fiber.Ctx) error {
	// body parse the data
	// check in stock_balances for that user
	// then check the symbol for that user
	// then check the yes or not quantity for that user
	// then deduct the quantity and lock the quantity
	// check for the symbol
	// check for price if not then create new
	// manage total
	// manage orders add user id and quantity in orders
	type inputFormat struct {
		UserId      string  `json:"userId"`
		StockSymbol string  `json:"stockSymbol"`
		Quantity    int     `json:"quantity"`
		Price       float32 `json:"price"`
		StockType   string  `json:"stockType"`
	}

	var inputData inputFormat
	err := c.BodyParser(&inputData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}

	userStockBalances, userExists := data.STOCK_BALANCES[inputData.UserId]
	if !userExists {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	stockBalance, stockExists := userStockBalances[inputData.StockSymbol]
	if !stockExists {
		return c.Status(fiber.StatusNotFound).SendString("Symbol not found")
	}

	getAvailableQuantity := func(stockType string) (int, error) {
		switch stockType {
		case "yes":
			return stockBalance.Yes.Quantity, nil
		case "no":
			return stockBalance.No.Quantity, nil
		default:
			return 0, fmt.Errorf("invalid stock type")
		}
	}

	availableQuantity, err := getAvailableQuantity(inputData.StockType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if availableQuantity < inputData.Quantity {
		return c.Status(fiber.StatusBadRequest).SendString("Insufficient balance")
	}

	switch inputData.StockType {
	case "yes":
		stockBalance.Yes.Quantity -= inputData.Quantity
	case "no":
		stockBalance.No.Quantity -= inputData.Quantity
	}
	data.STOCK_BALANCES[inputData.UserId][inputData.StockSymbol] = stockBalance

	updateOrderBook := func(orderBook data.OrderYesNo) data.OrderYesNo {
		strPrice := fmt.Sprintf("%.2f", inputData.Price)
		priceOption, exists := orderBook[strPrice]
		if !exists {
			priceOption = data.PriceOptions{
				Total:  0,
				Orders: make(data.Order),
			}
		}

		priceOption.Total += inputData.Quantity
		// priceOption.Orders[inputData.UserId].Quantity += inputData.Quantity

		orderBook[strPrice] = priceOption
		return orderBook
	}

	stockData, exists := data.ORDERBOOK[inputData.StockSymbol]
	if !exists {
		stockData = data.OrderSymbol{
			Yes: make(data.OrderYesNo),
			No:  make(data.OrderYesNo),
		}
	}

	switch inputData.StockType {
	case "yes":
		stockData.Yes = updateOrderBook(stockData.Yes)
	case "no":
		stockData.No = updateOrderBook(stockData.No)
	}

	data.ORDERBOOK[inputData.StockSymbol] = stockData

	return c.SendString(fmt.Sprintf("Sell order placed for %v '%s' options at price %v.", inputData.Quantity, inputData.StockType, inputData.Price))
}

// orderbook
//check the symbol
// if found then check for prices
// if price is <= then found prices and then update the oderbook and stock_balances
// if price is found partially then you have to update the orderbook and stock_balances then also work with reverse order for rest of the quantity
// if price is not found > then create the new price on the reverse side whith same quantity and set the user equals to userid itself and also give the flag or something for reverse order

// if symbol not found then create the new symbol along with price and other details in reverse order

// update inr balance
// balnce lock

// func BuyOrder(c *fiber.Ctx) error {
// 	type inputFormat struct {
// 		UserId      string  `json:"userId"`
// 		StockSymbol string  `json:"stockSymbol"`
// 		Quantity    int     `json:"quantity"`
// 		Price       float64 `json:"price"`
// 		StockType   string  `json:"stockType"`
// 	}

// 	var inputData inputFormat
// 	err := c.BodyParser(&inputData)

// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
// 	}

// 	// inr balances check
// 	userInr, exist := global.UserManager.INR_BALANCES[inputData.UserId]

// 	if !exist {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User not found"})
// 	}
// 	totalCost := float32(inputData.Price) * float32(inputData.Quantity)
// 	if userInr.Balance < totalCost {
// 		return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"message": "Insufficient balance"})
// 	}

// 	userInr.Locked += totalCost
// 	userInr.Balance -= totalCost
// 	global.UserManager.INR_BALANCES[inputData.UserId] = userInr

// 	createReverseOrder := func(inputData inputFormat) error {
// 		inputPrice := inputData.Price
// 		reversePrice := 10 - inputPrice
// 		reverseType := "no"
// 		if inputData.StockType == "no" {
// 			reverseType = "yes"
// 		}

// 		availableSymbol, exist := data.ORDERBOOK[inputData.StockSymbol]
// 		if !exist {
// 			availableSymbol = data.OrderSymbol{
// 				Yes: make(data.OrderYesNo),
// 				No:  make(data.OrderYesNo),
// 			}
// 			data.ORDERBOOK[inputData.StockSymbol] = availableSymbol
// 		}
// 		priceStr := strconv.FormatFloat(reversePrice, 'f', 2, 64)
// 		var reverseOrders data.PriceOptions
// 		if reverseType == "yes" {
// 			reverseOrders, exist = availableSymbol.Yes[priceStr]
// 		} else {
// 			reverseOrders, exist = availableSymbol.No[priceStr]
// 		}

// 		if !exist {
// 			reverseOrders = data.PriceOptions{
// 				Total:  inputData.Quantity,
// 				Orders: make(map[string]int),
// 			}
// 		} else {
// 			reverseOrders.Total += inputData.Quantity
// 		}
// 		reverseOrders.Orders[inputData.UserId] = inputData.Quantity

// 		if reverseType == "yes" {
// 			availableSymbol.Yes[priceStr] = reverseOrders
// 		} else {
// 			availableSymbol.No[priceStr] = reverseOrders
// 		}

// 		return nil
// 	}

// 	//orderbook
// 	availableSymbol, exist := data.ORDERBOOK[inputData.StockSymbol]
// 	if !exist {
// 		// reverse order
// 		createReverseOrder(inputData)
// 		return c.JSON(fiber.Map{"message": "Buy order placed and trade executed"})
// 	}

// 	var keyList []string = common.GetMapKeys(availableSymbol)
// 	sort.Strings(keyList)

// 	for _, v := range keyList {
// 		// lowestPrice := strconv.FormatFloat(v, 'g', 2, 32)
// 		lowestPrice, err := strconv.ParseFloat(v, 64)
// 		if err != nil {
// 			return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
// 		}

// 		if lowestPrice > inputData.Price {
// 			// handle reverse order
// 			createReverseOrder(inputData)
// 			return c.JSON(fiber.Map{"message": "Buy order placed and trade executed"})
// 		}
// 		if inputData.StockType == "yes" {
// 			currOrders, exist := availableSymbol.Yes[v]
// 			if !exist {
// 				continue
// 			}
// 			for userId, orderval := range currOrders.Orders {
// 				// update stock_balances function
// 				if inputData.Quantity >= orderval {
// 					inputData.Quantity -= orderval
// 					currOrders.Orders[userId] = 0
// 					delete(currOrders.Orders, userId)
// 				} else {
// 					currOrders.Orders[userId] -= inputData.Quantity
// 					if currOrders.Orders[userId] == 0 {
// 						delete(currOrders.Orders, userId)
// 					}
// 					inputData.Quantity = 0
// 					break
// 				}
// 			}
// 			availableSymbol.Yes[v] = currOrders
// 		}

// 	}

// 	return c.SendString("Buy order placed and trade executed")
// }

func createReverseOrder(stockSymbol, stockType, userId string, price float64, quantity int) {
	reversePrice := 1000 - price
	reverseType := "no"
	if stockType == "no" {
		reverseType = "yes"
	}

	availableSymbol, exists := data.ORDERBOOK[stockSymbol]
	if !exists {
		availableSymbol = data.OrderSymbol{
			Yes: make(data.OrderYesNo),
			No:  make(data.OrderYesNo),
		}
		data.ORDERBOOK[stockSymbol] = availableSymbol
	}

	priceStr := strconv.FormatFloat(reversePrice, 'f', 2, 64)

	//check
	// less than or equal to input price
	var reverseOrders data.PriceOptions
	if reverseType == "yes" {
		reverseOrders, exists = availableSymbol.Yes[priceStr]
	} else {
		reverseOrders, exists = availableSymbol.No[priceStr]
	}

	if !exists {
		reverseOrders = data.PriceOptions{
			Total:  quantity,
			Orders: make(map[string]data.OrderOptions),
		}
	} else {
		reverseOrders.Total += quantity
	}

	reverseOrders.Orders[userId] = data.OrderOptions{
		Quantity: quantity,
		Reverse:  true,
	}

	if reverseType == "yes" {
		availableSymbol.Yes[priceStr] = reverseOrders
	} else {
		availableSymbol.No[priceStr] = reverseOrders
	}

	data.ORDERBOOK[stockSymbol] = availableSymbol
}

func checkAndLockBalance(userId string, price int, quantity int) (bool, error) {
	user, exists := global.UserManager.GetUser(userId)
	if !exists {
		return false, fmt.Errorf("User not found")
	}

	totalCost := price * quantity

	if user.Balance < totalCost {
		return false, fmt.Errorf("Insufficient balance")
	}
	leftBalance := user.Balance - totalCost
	lockedAmount := user.Locked + totalCost

	global.UserManager.UpdateUserInrBalance(userId, leftBalance)
	global.UserManager.UpdateUserInrLock(userId, lockedAmount)

	return true, nil
}

type inputFormat struct {
	UserId      string `json:"userId"`
	StockSymbol string `json:"stockSymbol"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
	StockType   string `json:"stockType"`
}

func BuyOrder(c *fiber.Ctx) error {

	var inputData inputFormat
	if err := c.BodyParser(&inputData); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input data")
	}

	// Check and lock balance
	ok, err := checkAndLockBalance(inputData.UserId, inputData.Price, inputData.Quantity)
	if !ok {
		// failedBuyOrder(inputData)
		if err.Error() == "User not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"message": err.Error()})
	}

	availableSymbol, exists := global.OrderBookManager.GetOrderBook(inputData.StockSymbol)
	if !exists {

		// create symbol in orderbook
		global.OrderBookManager.AddOrderBookSymbol(inputData.StockSymbol)

		// check it can place order or not
		canPlace := CheckCanPlaceOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType)

		switch canPlace {
		case "fulfill":
			PlaceFullfillOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		case "partial":
			PlacePartialOrder()
		case "none":
			PlaceReverseBuyOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		}
		// send this event to redis queue with symbol
		// createReverseOrder(inputData.StockSymbol, inputData.StockType, inputData.UserId, inputData.Price, inputData.Quantity)
		return c.JSON(fiber.Map{"message": "Buy order placed and reverse order created"})
	}
	//not correct
	// it can be no or yes
	var orderBook data.OrderYesNo
	if inputData.StockType == "yes" {
		orderBook = availableSymbol.Yes
	} else if inputData.StockType == "no" {
		orderBook = availableSymbol.No
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid stock type"})
	}
	keyList := common.GetMapKeys(orderBook)
	sort.Strings(keyList)

	for _, priceKey := range keyList {
		lowestPrice, err := strconv.ParseFloat(priceKey, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid price in order book")
		}

		if lowestPrice > inputData.Price {
			createReverseOrder(inputData.StockSymbol, inputData.StockType, inputData.UserId, inputData.Price, inputData.Quantity)
			return c.JSON(fiber.Map{"message": "Buy order placed and reverse order created"})
		}

		// Handle matching price orders for "yes" stock type
		// if inputData.StockType == "yes" {
		// 	currOrders := availableSymbol.Yes[priceKey]
		// 	for userId, orderVal := range currOrders.Orders {
		// 		if inputData.Quantity >= orderVal.Quantity {
		// 			inputData.Quantity -= orderVal.Quantity
		// 			delete(currOrders.Orders, userId)
		// 		} else {
		// 			currOrders.Orders[userId].Quantity -= inputData.Quantity
		// 			inputData.Quantity = 0
		// 			break
		// 		}
		// 	}
		// 	availableSymbol.Yes[priceKey] = currOrders

		// 	// Exit early if all requested quantity is fulfilled
		// 	if inputData.Quantity == 0 {
		// 		break
		// 	}
		// }

		priceOptions := orderBook[priceKey]
		for userId, orderOptions := range priceOptions.Orders {
			// If the buy quantity is greater or equal to the sell order quantity
			if inputData.Quantity >= orderOptions.Quantity {
				// Fulfill this order, reduce the buy quantity, and remove the sell order
				inputData.Quantity -= orderOptions.Quantity
				delete(priceOptions.Orders, userId)
			} else {
				// Partially fulfill the order, reduce the sell order quantity
				priceOptions.Orders[userId] = data.OrderOptions{
					Quantity: orderOptions.Quantity - inputData.Quantity,
					Reverse:  orderOptions.Reverse,
				}
				inputData.Quantity = 0
				break
			}
		}

		// Update the order book for this price level
		orderBook[priceKey] = priceOptions

		// If the entire buy quantity has been fulfilled, exit early
		if inputData.Quantity == 0 {
			break
		}
	}

	// If there is remaining quantity, create a reverse order
	if inputData.Quantity > 0 {
		createReverseOrder(inputData.StockSymbol, inputData.StockType, inputData.UserId, inputData.Price, inputData.Quantity)
	}

	return c.JSON(fiber.Map{"message": "Buy order placed and trade executed"})
}
