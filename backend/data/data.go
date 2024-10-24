// data/managers.go
package data

import (
	"fmt"
	"sync"
)

// User balance structures
type User struct {
	Balance float32 `json:"balance"`
	Locked  float32 `json:"locked"`
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

// Thread-safe methods for UserManager
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
