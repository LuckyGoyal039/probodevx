package data

type User struct {
	Balance float32
	Locked  float32
}

var INR_BALANCES = make(map[string]User)

type Order struct {
	Stockname map[string]interface{}
}

var ORDERBOOK = Order{}

// "BTC_USDT_10_Oct_2024_9_30": {
// 		 "yes": {
// 			 "9.5": {
// 				 "total": 12,
// 				 orders: {
// 					 "user1": 2,
// 					 "user2": 10
// 				 }
// 			 },
// 			 "8.5": {
// 				 "total": 12,
// 				 "orders": {
// 					 "user1": 3,
// 					 "user2": 3,
// 					 "user3": 6
// 				 }
// 			 },
// 		 },
// 		 "no": {

// 		 }
// }

type StockBal struct {
	UserId map[string]interface{}
}

var STOCK_BALANCES = StockBal{}

// user1: {
//    "BTC_USDT_10_Oct_2024_9_30": {
// 	   "yes": {
// 		   "quantity": 1,
// 		   "locked": 0
// 	   }
//    }
// },
// user2: {
// 	"BTC_USDT_10_Oct_2024_9_30": {
// 	   "no": {
// 		   "quantity": 3,
// 		   "locked": 4
// 	   }
//    }
// }
