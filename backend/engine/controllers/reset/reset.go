package reset

import (
	"context"
	"fmt"

	"github.com/probodevx/engine/data"
	"github.com/probodevx/engine/global"
	"github.com/probodevx/engine/shared"
)

func ResetAll(ctx context.Context, event shared.EventModel) (interface{}, error) {

	if ok := data.ResetAllManager(global.UserManager, global.StockManager, global.OrderBookManager); !ok {
		return nil, fmt.Errorf("something went wrong")
	}
	return map[string]interface{}{
		"message": "reset successfully",
	}, nil
}
