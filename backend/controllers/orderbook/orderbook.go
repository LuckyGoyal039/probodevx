package orderbook

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/probodevx/global"
)

func GetOrderbookSymbol(c *fiber.Ctx) error {
	stockSymbol := utils.CopyString(c.Params("stockSymbol"))
	if stockSymbol == "" {
		return c.JSON(global.OrderBookManager.GetAllOrderBook())
	}
	newData, exists := global.OrderBookManager.GetOrderBook(stockSymbol)

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("stock symbol not found")
	}
	return c.JSON(newData)
}

func SellOrder(c *fiber.Ctx) error {
	type inputFormat struct {
		UserId      string `json:"userId"`
		StockSymbol string `json:"stockSymbol"`
		Quantity    int    `json:"quantity"`
		Price       int    `json:"price"`
		StockType   string `json:"stockType"`
	}

	var inputData inputFormat
	err := c.BodyParser(&inputData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}
	ok := CheckAndLockStock(inputData.UserId, inputData.StockSymbol, inputData.StockType, inputData.Quantity)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Insufficient stock balance"})
	}

	canPlace := CheckBuyer(inputData.StockSymbol, inputData.StockType, inputData.Price, inputData.Quantity)
	switch canPlace {
	case "fullfill":
		FullFillSellOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		// UnLockBalance(inputData.UserId, inputData.Quantity)
	case "partial":
		PlacePartialSellOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		// updatedLockedAmount := inputData.Quantity - GetFullFillableQuantity(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType)
		// UnLockBalance(inputData.UserId, updatedLockedAmount)
	case "none":
		PlaceSellOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		// UnLockBalance(inputData.UserId, inputData.Quantity)
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Sell order placed for %v '%s' options at price %v.", inputData.Quantity, inputData.StockType, inputData.Price),
	})
}

type inputFormat struct {
	UserId      string `json:"userId"`
	StockSymbol string `json:"stockSymbol"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
	StockType   string `json:"stockType"`
}

// func BuyOrder(c *fiber.Ctx) error {

// 	var inputData inputFormat
// 	if err := c.BodyParser(&inputData); err != nil {
// 		return c.Status(fiber.StatusBadRequest).SendString("Invalid input data")
// 	}
// 	ok, err := checkAndLockBalance(inputData.UserId, inputData.Price, inputData.Quantity)
// 	if !ok {
// 		if err.Error() == "User not found" {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
// 		}
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
// 	}

// 	// check it can place order or not
// 	canPlace := CheckCanPlaceOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType)

// 	switch canPlace {
// 	case "fullfill":
// 		//check this out
// 		PlaceFullFillOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
// 		UnLockBalance(inputData.UserId, inputData.Quantity, inputData.Price)
// 		// 	balance, _ := global.UserManager.GetUserBalance(inputData.UserId)
// 		// 	balance -= inputData.Price * inputData.Quantity
// 		// 	global.UserManager.UpdateUserInrBalance(inputData.UserId, balance)
// 		// case "partial":
// 		PlacePartialOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
// 		updatedLockedAmount := inputData.Quantity - GetFullFillableQuantity(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType)
// 		UnLockBalance(inputData.UserId, updatedLockedAmount, inputData.Price)
// 		// balance, _ := global.UserManager.GetUserBalance(inputData.UserId)
// 		// balance -= inputData.Price * updatedLockedAmount
// 		// global.UserManager.UpdateUserInrBalance(inputData.UserId, balance)
// 	case "none":
// 		PlaceReverseBuyOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
// 		// UnLockBalance(inputData.UserId, inputData.Quantity, inputData.Price)
// 	}
// 	if orderbookData, exists := global.OrderBookManager.GetOrderBook(inputData.StockSymbol); exists {
// 		err := PushInQueue(inputData.StockSymbol, orderbookData)
// 		if err != nil {
// 			panic(fmt.Sprintf("error: %s", err))
// 		}
// 	}
// 	return c.JSON(fiber.Map{"message": "Buy order placed and trade executed"})
// }

func BuyOrder(c *fiber.Ctx) error {
	var inputData inputFormat
	if err := c.BodyParser(&inputData); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input data")
	}
	if ok := checkValidBalance(inputData.UserId, inputData.Price, inputData.Quantity); !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Insufficient INR balance"})
	}
	priceList := GetValidPrices(inputData.StockSymbol, inputData.StockType, inputData.Price)
	remainingQuantity := inputData.Quantity
	for _, currentPrice := range priceList {
		priceStr := strconv.FormatInt(int64(currentPrice), 10)
		priceData := global.OrderBookManager.GetPriceMap(inputData.StockSymbol, inputData.StockType, priceStr)

		for sellerId, orderInfo := range priceData.Orders {
			if remainingQuantity <= 0 {
				break
			}
			availableQuantity := orderInfo.Quantity
			quantityToTake := min(remainingQuantity, availableQuantity)

			if quantityToTake > 0 {

				amount := quantityToTake * inputData.Price
				global.UserManager.DebitBalance(inputData.UserId, amount) //debit the qty of user
				global.UserManager.CreditBalance(sellerId, amount)        //credit the quantity of seller

				//remove lock of seller in stock_balances
				lockedQty, _ := global.StockManager.GetLockedStocks(sellerId, inputData.StockSymbol, inputData.StockType)
				lockedQty -= quantityToTake
				global.StockManager.SetStocksLock(sellerId, inputData.StockSymbol, inputData.StockType, lockedQty)

				// Update buyer's stock balance
				AddStocksToBuyer(inputData.UserId, inputData.StockSymbol, inputData.StockType, inputData.Quantity)

				// update total and quantity of seller
				priceData.Total -= quantityToTake
				orderInfo.Quantity -= quantityToTake
				if orderInfo.Quantity == 0 {
					delete(priceData.Orders, sellerId)
				}

				remainingQuantity -= quantityToTake
			}
		}
	}
	if remainingQuantity > 0 {
		PlaceReverseBuyOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
	}
	if orderbookData, exists := global.OrderBookManager.GetOrderBook(inputData.StockSymbol); exists {
		err := PushInQueue(inputData.StockSymbol, orderbookData)
		if err != nil {
			panic(fmt.Sprintf("error: %s", err))
		}
	}
	return c.JSON(fiber.Map{"message": "Buy order placed and trade executed"})
}
