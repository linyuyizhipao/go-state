package main

import (
	"fmt"
	"slp/test/state/database"
	"slp/test/state/order"
)

func main() {
	order.Init()
	orderList, dbErr := database.ListAllOrder()
	if dbErr != nil {
		return
	}
	for _, curOrder := range orderList {
		params := make(map[string]interface{})
		params["order"] = curOrder
		if err := order.ExecOrderTask(params); err != nil {
			fmt.Printf("执行订单任务出错：%v\n", err)
		}
		fmt.Println("\n\n")
	}
}
