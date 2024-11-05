package inrBalance

import (
	"context"
	"fmt"

	"github.com/probodevx/engine/global"
	"github.com/probodevx/engine/shared"
)

func GetInrBalance(ctx context.Context, event shared.EventModel) (interface{}, error) {
	userId := event.UserId

	if userId == "" {
		return global.UserManager.GetAllUsers(), nil
	}
	newData, exists := global.UserManager.GetUser(userId)

	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return newData, nil
}

func AddUserBalance(ctx context.Context, event shared.EventModel) (interface{}, error) {
	userId := event.UserId
	amountFloat, ok := event.Data["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount format")
	}
	amount := int64(amountFloat)

	// Retrieve user and update balance as before
	userData, exists := global.UserManager.GetUser(userId)
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	newBalance := userData.Balance + int(amount)
	if _, err := global.UserManager.UpdateUserInrBalance(userId, newBalance); err != nil {
		return nil, fmt.Errorf("failed to update user balance: %v", err)
	}

	return map[string]interface{}{
		"userId":  userId,
		"balance": newBalance,
		"message": fmt.Sprintf("Onramped %s with amount %v", userId, newBalance),
	}, nil
}
