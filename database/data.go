package database

import "fmt"

// DB 模拟数据库对象
type DB struct {
}

// Transaction 模拟事务
func (db *DB) Transaction(fun func() error) error {
	fmt.Println("事务执行开始。")
	err := fun()
	fmt.Println("事务执行结束。")
	return err
}

// Order 订单
type Order struct {
	ID    int64 // 主键ID
	State int   // 状态
}

type OrderList []*Order

// 查询所有订单
func ListAllOrder() (OrderList, error) {
	orderList := OrderList{
		&Order{1, 0},
		&Order{2, 1},
		&Order{2, 2},
	}
	return orderList, nil
}

// UpdateOrderState 更新订单状态
func UpdateOrderState(curOrder *Order, srcState int, dstState int) error {
	if curOrder.State == srcState {
		curOrder.State = dstState
	}
	fmt.Printf("更新id为 %v 的订单状态，从现态[%v]到次态[%v]\n", curOrder.ID, srcState, dstState)
	return nil
}
