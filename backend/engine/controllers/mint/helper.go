package mint

import (
	"fmt"
	"github.com/probodevx/engine/global"
)

func updateBalance(price int, quantity int, userId string) error {
	user, exist := global.UserManager.GetUser(userId)
	if !exist {
		return fmt.Errorf("user not found")
	}
	currentBalance := user.Balance
	requiredBalance := price * quantity
	if currentBalance < requiredBalance {
		return fmt.Errorf("insufficient balance")
	}
	newBalance := currentBalance - requiredBalance
	global.UserManager.UpdateUserInrBalance(userId, newBalance)
	return nil
}
