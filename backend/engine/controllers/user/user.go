package user

import (
	"context"
	"fmt"

	"github.com/probodevx/engine/global"
	"github.com/probodevx/engine/shared"
)

func CreateNewUser(ctx context.Context, event shared.UserEvent) (interface{}, error) {
	userId := event.UserId

	if userId == "" {
		return nil, fmt.Errorf("invalid userId")
	}

	if err := global.UserManager.CreateUser(userId); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("User %s created", userId),
	}, nil
}
