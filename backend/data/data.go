// data/managers.go
package data

import (
	"fmt"
	"strconv"
	"sync"
)

// inr balances
type User struct {
	Balance int `json:"balance"`
	Locked  int `json:"locked"`
}

type UserManager struct {
	mu          sync.Mutex
	inrBalances map[string]*User
}

func NewUserManager() *UserManager {
	return &UserManager{
		inrBalances: make(map[string]*User),
	}
}

// orderbook
type OrderOptions struct {
	Quantity int  `json:"quantity"`
	Reverse  bool `json:"reverse"`
}

type Order map[string]OrderOptions
type PriceOptions struct {
	Total  int   `json:"total"`
	Orders Order `json:"orders"`
}
type OrderYesNo map[string]PriceOptions
type OrderSymbol struct {
	Yes OrderYesNo `json:"yes"`
	No  OrderYesNo `json:"no"`
}

type OrderBookManager struct {
	mu        sync.Mutex
	orderBook map[string]OrderSymbol
}

func NewOrderBookManager() *OrderBookManager {
	return &OrderBookManager{
		orderBook: make(map[string]OrderSymbol),
	}
}

// stock balances
type YesNo struct {
	Quantity int `json:"quantity"`
	Locked   int `json:"locked"`
}

type StockOption struct {
	Yes YesNo `json:"yes"`
	No  YesNo `json:"no"`
}

type UserStockBalance map[string]StockOption

type StockManager struct {
	mu            sync.Mutex
	stockBalances map[string]UserStockBalance
}

func NewStockManager() *StockManager {
	return &StockManager{
		stockBalances: make(map[string]UserStockBalance),
	}
}

func (um *UserManager) CreateUser(userId string) error {
	um.mu.Lock()
	um.mu.Unlock()

	if _, exists := um.inrBalances[userId]; exists {
		return fmt.Errorf("user already exists")
	}

	um.inrBalances[userId] = &User{
		Balance: 0,
		Locked:  0,
	}
	return nil
}
func (um *UserManager) GetUser(userId string) (*User, bool) {
	// um.mu.Lock()
	//  um.mu.Unlock()

	user, exists := um.inrBalances[userId]
	return user, exists
}
func (um *UserManager) GetUserBalance(userId string) (int, bool) {
	// um.mu.Lock()
	//  um.mu.Unlock()

	user, exists := um.inrBalances[userId]
	balance := user.Balance
	return balance, exists
}
func (um *UserManager) GetUserLocked(userId string) (int, bool) {
	// um.mu.Lock()
	//  um.mu.Unlock()

	user, exists := um.inrBalances[userId]
	Locked := user.Locked
	return Locked, exists
}

func (um *UserManager) GetAllUsers() map[string]User {
	// um.mu.Lock()
	//  um.mu.Unlock()

	// Create a copy to prevent external modifications
	result := make(map[string]User)
	for k, v := range um.inrBalances {
		result[k] = *v
	}
	return result
}
func (um *UserManager) UpdateUserInrBalance(userId string, balance int) (*User, error) {
	user, exists := um.inrBalances[userId]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	user.Balance = balance
	um.mu.Lock()
	um.inrBalances[userId] = user
	um.mu.Unlock()
	return user, nil
}
func (um *UserManager) DebitBalance(userId string, amount int) (*User, error) {
	user, exists := um.inrBalances[userId]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	user.Balance -= amount
	um.mu.Lock()
	um.inrBalances[userId] = user
	um.mu.Unlock()
	return user, nil
}
func (um *UserManager) CreditBalance(userId string, amount int) (*User, error) {
	user, exists := um.inrBalances[userId]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	user.Balance += amount
	um.mu.Lock()
	um.inrBalances[userId] = user
	um.mu.Unlock()
	return user, nil
}

func (um *UserManager) UpdateUserInrLock(userId string, lock int) (*User, error) {
	user, exists := um.inrBalances[userId]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	user.Locked = lock
	um.mu.Lock()
	um.inrBalances[userId] = user
	um.mu.Unlock()
	return user, nil
}

func (sm *StockManager) GetStockBalances(userId string) (UserStockBalance, bool) {
	balances, exists := sm.stockBalances[userId]
	return balances, exists
}
func (sm *StockManager) GetAllStockBalances() map[string]UserStockBalance {
	result := make(map[string]UserStockBalance)

	for key, value := range sm.stockBalances {
		result[key] = value
	}
	return result
}
func (sm *StockManager) AddStockBalancesSymbol(stockSymbol string) {

	for userID, stockBalance := range sm.stockBalances {
		if _, exists := stockBalance[stockSymbol]; !exists {
			stockBalance[stockSymbol] = StockOption{
				Yes: YesNo{
					Quantity: 0,
					Locked:   0,
				},
				No: YesNo{
					Quantity: 0,
					Locked:   0,
				},
			}
			sm.mu.Lock()
			sm.stockBalances[userID] = stockBalance
			sm.mu.Unlock()
		}
	}
}
func (sm *StockManager) UpdateStockBalanceSymbol(userId string, stockSymbol string, data StockOption) (UserStockBalance, error) {

	user, exists := sm.stockBalances[userId]
	if !exists {
		return UserStockBalance{}, fmt.Errorf("User not found")
	}
	user[stockSymbol] = data
	sm.mu.Lock()
	sm.stockBalances[userId] = user
	sm.mu.Unlock()
	return user, nil
}
func (sm *StockManager) CheckUser(userId string) bool {
	if _, exist := sm.stockBalances[userId]; exist {
		return true
	}
	return false
}
func (sm *StockManager) AddNewUser(userId string) (UserStockBalance, error) {
	if exist := sm.CheckUser(userId); exist {
		return UserStockBalance{}, fmt.Errorf("user already exist")
	}
	sm.mu.Lock()
	sm.stockBalances[userId] = UserStockBalance{}
	sm.mu.Unlock()
	return sm.stockBalances[userId], nil
}
func (sm *StockManager) CheckSymbolForUser(userId string, stockSymbol string) bool {
	data, exist := sm.GetStockBalances(userId)
	if !exist {
		return false
	}
	if _, exists := data[stockSymbol]; !exists {
		return false
	}
	return true
}
func (sm *StockManager) SetStocksLock(userId string, stockSymbol string, stockType string, quantity int) error {
	if exist := sm.CheckUser(userId); !exist {
		return fmt.Errorf("user not found")
	}
	user, _ := sm.GetStockBalances(userId)

	stockData, exists := user[stockSymbol]
	if !exists {
		return fmt.Errorf("stock symbol not found")
	}
	if stockType == "yes" {
		stockData.Yes.Locked = quantity
	} else if stockType == "no" {
		stockData.No.Locked = quantity
	}
	sm.stockBalances[userId][stockSymbol] = stockData
	return nil
}

// func (sm *StockManager) UnLockStocks(userId string, stockSymbol string, stockType string, quantity int) error {
// 	if exist := sm.CheckUser(userId); !exist {
// 		return fmt.Errorf("user not found")
// 	}
// 	user, _ := sm.GetStockBalances(userId)

//		stockData, exists := user[stockSymbol]
//		if !exists {
//			return fmt.Errorf("stock symbol not found")
//		}
//		if stockType == "yes" {
//			stockData.Yes.Locked -= quantity
//		} else if stockType == "no" {
//			stockData.No.Locked -= quantity
//		}
//		sm.stockBalances[userId][stockSymbol] = stockData
//		return nil
//	}
func (sm *StockManager) SetStocksQuantity(userId string, stockSymbol string, stockType string, quantity int) error {
	if exist := sm.CheckUser(userId); !exist {
		return fmt.Errorf("user not found")
	}
	user, _ := sm.GetStockBalances(userId)

	stockData, exists := user[stockSymbol]
	if !exists {
		return fmt.Errorf("stock symbol not found")
	}
	if stockType == "yes" {
		stockData.Yes.Quantity = quantity
	} else if stockType == "no" {
		stockData.No.Quantity = quantity
	}
	sm.stockBalances[userId][stockSymbol] = stockData
	return nil
}
func (sm *StockManager) GetLockedStocks(userId string, stockSymbol string, stockType string) (int, error) {
	if exist := sm.CheckUser(userId); !exist {
		return 0, fmt.Errorf("user not found")
	}
	user, _ := sm.GetStockBalances(userId)

	stockData, exists := user[stockSymbol]
	if !exists {
		return 0, fmt.Errorf("stock symbol not found")
	}
	var lockedAmount int
	if stockType == "yes" {
		lockedAmount = stockData.Yes.Locked
	} else if stockType == "no" {
		lockedAmount = stockData.No.Locked
	}
	return lockedAmount, nil
}
func (sm *StockManager) GetQuantityStocks(userId string, stockSymbol string, stockType string) (int, error) {
	if exist := sm.CheckUser(userId); !exist {
		return 0, fmt.Errorf("user not found")
	}
	user, _ := sm.GetStockBalances(userId)

	stockData, exists := user[stockSymbol]
	if !exists {
		return 0, fmt.Errorf("stock symbol not found")
	}
	var quantityAmount int
	if stockType == "yes" {
		quantityAmount = stockData.Yes.Quantity
	} else if stockType == "no" {
		quantityAmount = stockData.No.Quantity
	}
	return quantityAmount, nil
}

func (om *OrderBookManager) GetOrderBook(stockSymbol string) (OrderSymbol, bool) {

	// om.mu.Lock()
	//  om.mu.Unlock()
	symbol, exists := om.orderBook[stockSymbol]
	return symbol, exists
}
func (om *OrderBookManager) GetAllOrderBook() map[string]OrderSymbol {

	// om.mu.Lock()
	//  om.mu.Unlock()
	result := om.orderBook
	for k, v := range om.orderBook {
		result[k] = v
	}
	return result
}
func (om *OrderBookManager) AddOrderBookSymbol(stockSymbol string) OrderSymbol {
	var newSymbol = OrderSymbol{
		Yes: make(OrderYesNo),
		No:  make(OrderYesNo),
	}
	om.mu.Lock()
	om.orderBook[stockSymbol] = newSymbol
	om.mu.Unlock()
	return newSymbol
}
func (om *OrderBookManager) CreateOrderbookPrice(stockSymbol string, stockType string, price int, quantity int, userId string, reverse bool) {
	var orderData OrderYesNo
	orderSymbol, exists := om.GetOrderBook(stockSymbol)
	if !exists {
		orderSymbol = OrderSymbol{
			Yes: make(OrderYesNo),
			No:  make(OrderYesNo),
		}
		om.AddOrderBookSymbol(stockSymbol)
	}

	if stockType == "yes" {
		orderData = orderSymbol.Yes
	} else if stockType == "no" {
		orderData = orderSymbol.No
	}

	priceStr := strconv.FormatInt(int64(price), 10)

	priceLevel, exists := orderData[priceStr]
	if !exists {
		priceLevel = PriceOptions{
			Total:  quantity,
			Orders: make(Order),
		}
	} else {
		priceLevel.Total += quantity
	}

	if userOrder, exists := priceLevel.Orders[userId]; exists {
		userOrder.Quantity += quantity
		userOrder.Reverse = reverse
		priceLevel.Orders[userId] = userOrder
	} else {
		priceLevel.Orders[userId] = OrderOptions{
			Quantity: quantity,
			Reverse:  reverse,
		}
	}

	orderData[priceStr] = priceLevel
	if stockType == "yes" {
		orderSymbol.Yes = orderData
	} else {
		orderSymbol.No = orderData
	}
	om.mu.Lock()
	om.orderBook[stockSymbol] = orderSymbol
	om.mu.Unlock()
}

func (om *OrderBookManager) CheckStockSymbol(stockSymbol string) bool {
	if _, exists := om.orderBook[stockSymbol]; !exists {
		return false
	}
	return true
}
func (om *OrderBookManager) RemoveUserFromOrder(userId string, stockSymbol string, stockType string, price int) {
	orderBook, exists := om.orderBook[stockSymbol]
	if !exists {
		return
	}

	var priceMap OrderYesNo
	switch stockType {
	case "yes":
		priceMap = orderBook.Yes
	case "no":
		priceMap = orderBook.No
	default:
		return
	}

	priceKey := fmt.Sprintf("%d", price)
	priceLevel, priceExists := priceMap[priceKey]
	if !priceExists {
		return
	}

	userOrder, userExists := priceLevel.Orders[userId]
	if !userExists {
		return
	}

	priceLevel.Total -= userOrder.Quantity
	delete(priceLevel.Orders, userId)

	if len(priceLevel.Orders) == 0 || priceLevel.Total <= 0 {
		delete(priceMap, priceKey)
	} else {
		priceMap[priceKey] = priceLevel
	}

	if stockType == "yes" {
		orderBook.Yes = priceMap
	} else {
		orderBook.No = priceMap
	}

	om.orderBook[stockSymbol] = orderBook
}
func (om *OrderBookManager) UpdateStockQtyFromOrder(userId string, stockSymbol string, stockType string, price, quantity int) {
	orderBook, exists := om.orderBook[stockSymbol]
	if !exists {
		return
	}

	var priceMap OrderYesNo
	switch stockType {
	case "yes":
		priceMap = orderBook.Yes
	case "no":
		priceMap = orderBook.No
	default:
		return
	}

	priceKey := fmt.Sprintf("%d", price)
	priceLevel, priceExists := priceMap[priceKey]
	if !priceExists {
		return
	}

	userOrder, userExists := priceLevel.Orders[userId]
	if !userExists {
		return
	}

	priceLevel.Total -= quantity
	userOrder.Quantity -= quantity
	priceLevel.Orders[userId] = userOrder

	if priceLevel.Total <= 0 {
		delete(priceMap, priceKey)
	} else {
		priceMap[priceKey] = priceLevel
	}

	if stockType == "yes" {
		orderBook.Yes = priceMap
	} else {
		orderBook.No = priceMap
	}

	om.orderBook[stockSymbol] = orderBook
}

func (om *OrderBookManager) AddOrderbookPrice(stockSymbol, stockType string, price int) PriceOptions {

	orderSymbol, exists := om.orderBook[stockSymbol]
	if !exists {
		orderSymbol = OrderSymbol{
			Yes: make(OrderYesNo),
			No:  make(OrderYesNo),
		}
		om.orderBook[stockSymbol] = orderSymbol
	}

	var priceMap OrderYesNo
	if stockType == "yes" {
		priceMap = orderSymbol.Yes
	} else if stockType == "no" {
		priceMap = orderSymbol.No
	} else {
		return PriceOptions{} // Invalid stock type
	}

	priceKey := fmt.Sprintf("%d", price)
	priceOptions, exists := priceMap[priceKey]
	if !exists {
		priceOptions = PriceOptions{
			Total:  0,
			Orders: make(Order),
		}
		priceMap[priceKey] = priceOptions
	}

	if stockType == "yes" {
		orderSymbol.Yes = priceMap
	} else {
		orderSymbol.No = priceMap
	}
	om.mu.Lock()
	om.orderBook[stockSymbol] = orderSymbol
	om.mu.Unlock()

	return priceOptions
}
func (om *OrderBookManager) UpdateSellOrder(userId string, stockSymbol string, stockType string, price, quantity int) {
	om.mu.Lock()
	defer om.mu.Unlock()

	orderBook, exists := om.orderBook[stockSymbol]
	if !exists {
		orderBook = OrderSymbol{
			Yes: make(OrderYesNo),
			No:  make(OrderYesNo),
		}
		om.orderBook[stockSymbol] = orderBook
	}

	var priceMap OrderYesNo
	switch stockType {
	case "yes":
		priceMap = orderBook.Yes
	case "no":
		priceMap = orderBook.No
	default:
		return
	}

	priceKey := fmt.Sprintf("%d", price)
	priceLevel, priceExists := priceMap[priceKey]
	if !priceExists {
		priceLevel = PriceOptions{
			Total:  0,
			Orders: make(Order),
		}
	}

	userOrder, userExists := priceLevel.Orders[userId]
	if userExists {
		priceLevel.Total += quantity
		userOrder.Quantity += quantity
	} else {
		priceLevel.Total += quantity
		userOrder = OrderOptions{Quantity: quantity, Reverse: false}
	}
	priceLevel.Orders[userId] = userOrder
	priceMap[priceKey] = priceLevel

	if stockType == "yes" {
		orderBook.Yes = priceMap
	} else {
		orderBook.No = priceMap
	}

	om.orderBook[stockSymbol] = orderBook
}
func (om *OrderBookManager) GetPriceMap(stockSymbol, stockType, price string) PriceOptions {
	var priceData PriceOptions
	orderbook, _ := om.orderBook[stockSymbol]
	if stockType == "yes" {
		priceData = orderbook.Yes[price]
	} else if stockType == "no" {
		priceData = orderbook.No[price]

	}
	return priceData
}

func ResetAllManager(um *UserManager, sm *StockManager, om *OrderBookManager) bool {
	um.mu.Lock()
	um.inrBalances = make(map[string]*User)
	um.mu.Unlock()

	sm.mu.Lock()
	sm.stockBalances = make(map[string]UserStockBalance)
	sm.mu.Unlock()

	om.mu.Lock()
	om.orderBook = make(map[string]OrderSymbol)
	om.mu.Unlock()
	return true
}
