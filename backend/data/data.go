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
	defer um.mu.Unlock()

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
	um.mu.Lock()
	defer um.mu.Unlock()

	user, exists := um.inrBalances[userId]
	return user, exists
}
func (um *UserManager) GetAllUsers() map[string]User {
	um.mu.Lock()
	defer um.mu.Unlock()

	// Create a copy to prevent external modifications
	result := make(map[string]User)
	for k, v := range um.inrBalances {
		result[k] = *v
	}
	return result
}
func (um *UserManager) UpdateUserInrBalance(userId string, balance int) (*User, error) {
	um.mu.Lock()
	defer um.mu.Unlock()
	user, exists := um.inrBalances[userId]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	user.Balance = balance
	um.inrBalances[userId] = user
	return user, nil
}
func (um *UserManager) UpdateUserInrLock(userId string, lock int) (*User, error) {
	um.mu.Lock()
	defer um.mu.Unlock()
	user, exists := um.inrBalances[userId]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	user.Locked = lock
	um.inrBalances[userId] = user
	return user, nil
}

func (sm *StockManager) GetStockBalances(userId string) (UserStockBalance, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	balances, exists := sm.stockBalances[userId]
	return balances, exists
}
func (sm *StockManager) GetAllStockBalances() map[string]UserStockBalance {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	result := make(map[string]UserStockBalance)

	for key, value := range sm.stockBalances {
		result[key] = value
	}
	return result
}
func (sm *StockManager) AddStockBalancesSymbol(stockSymbol string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

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
			sm.stockBalances[userID] = stockBalance
		}
	}
}
func (sm *StockManager) UpdateStockBalanceSymbol(userId string, stockSymbol string, data StockOption) (UserStockBalance, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	user, exists := sm.stockBalances[userId]
	if !exists {
		return UserStockBalance{}, fmt.Errorf("User not found")
	}
	user[stockSymbol] = data
	sm.stockBalances[userId] = user
	return user, nil
}

func (om *OrderBookManager) GetOrderBook(stockSymbol string) (OrderSymbol, bool) {

	om.mu.Lock()
	defer om.mu.Unlock()
	symbol, exists := om.orderBook[stockSymbol]
	return symbol, exists
}
func (om *OrderBookManager) GetAllOrderBook() map[string]OrderSymbol {

	om.mu.Lock()
	defer om.mu.Unlock()
	result := om.orderBook
	for k, v := range om.orderBook {
		result[k] = v
	}
	return result
}
func (om *OrderBookManager) AddOrderBookSymbol(stockSymbol string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	var newSymbol = OrderSymbol{
		Yes: make(OrderYesNo),
		No:  make(OrderYesNo),
	}
	om.orderBook[stockSymbol] = newSymbol
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
	om.orderBook[stockSymbol] = orderSymbol
}
func (om *OrderBookManager) UpdateOrderBookSymbol(stockSymbol string, data OrderSymbol) {

}

func ResetAllManager(um *UserManager, sm *StockManager, om *OrderBookManager) {
	um.mu.Lock()
	sm.mu.Lock()
	om.mu.Lock()

	defer um.mu.Unlock()
	defer sm.mu.Unlock()
	defer om.mu.Unlock()

	um.inrBalances = make(map[string]*User)
	sm.stockBalances = make(map[string]UserStockBalance)
	om.orderBook = make(map[string]OrderSymbol)
}
