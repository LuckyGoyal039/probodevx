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
	// sm.mu.Lock()
	//  sm.mu.Unlock()

	balances, exists := sm.stockBalances[userId]
	return balances, exists
}
func (sm *StockManager) GetAllStockBalances() map[string]UserStockBalance {
	// sm.mu.Lock()
	//  sm.mu.Unlock()
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
func (sm *StockManager) CheckUser(userId string) (UserStockBalance, bool) {
	// sm.mu.Lock()
	//  sm.mu.Unlock()
	if user, exist := sm.stockBalances[userId]; exist {
		return user, true
	}
	return UserStockBalance{}, false
}
func (sm *StockManager) AddNewUser(userId string) (UserStockBalance, error) {
	if _, exist := sm.CheckUser(userId); exist {
		return UserStockBalance{}, fmt.Errorf("user already exist")
	}
	sm.mu.Lock()
	sm.stockBalances[userId] = UserStockBalance{}
	sm.mu.Unlock()
	return sm.stockBalances[userId], nil
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
func (om *OrderBookManager) AddOrderBookSymbol(stockSymbol string) {
	var newSymbol = OrderSymbol{
		Yes: make(OrderYesNo),
		No:  make(OrderYesNo),
	}
	om.mu.Lock()
	om.orderBook[stockSymbol] = newSymbol
	om.mu.Unlock()
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

func (om *OrderBookManager) UpdateOrderBookSymbol(stockSymbol string, data OrderSymbol) {

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
