package order

import (
	"fmt"
	"slp/test/state/database"
)

type PlaceAction struct {
}

// Before 事务前执行，业务上允许多次操作
func (receiver PlaceAction) Before(bizParams map[string]interface{}) error {
	fmt.Println("执行下单的Before方法。")
	return nil
}

// Execute 事务中执行，与状态转移在同一事务中
func (receiver PlaceAction) Execute(bizParams map[string]interface{}, tx *database.DB) error {
	fmt.Println("执行下单的Execute方法。")
	return nil
}

// After 事务后执行，业务上允许执行失败或未执行
func (receiver PlaceAction) After(bizParams map[string]interface{}) error {
	fmt.Println("执行下单的After方法。")
	return nil
}
