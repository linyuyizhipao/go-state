package order

import (
	"fmt"
	"slp/test/state/database"
	"slp/test/state/fsm"
)

var (
	// 状态
	StateOrderInit          = fsm.FSMState(0) // 初始状态
	StateOrderToBePaid      = fsm.FSMState(1) // 待支付
	StateOrderToBeDelivered = fsm.FSMState(2) // 待发货
	StateOrderCancel        = fsm.FSMState(3) // 订单取消
	StateOrderToBeReceived  = fsm.FSMState(4) // 待收货
	StateOrderDone          = fsm.FSMState(5) // 订单完成

	// 事件
	EventOrderPlace      = fsm.FSMEvent("EventOrderPlace")      // 下单
	EventOrderPay        = fsm.FSMEvent("EventOrderPay")        // 支付
	EventOrderPayTimeout = fsm.FSMEvent("EventOrderPayTimeout") // 支付超时
	EventOrderDeliver    = fsm.FSMEvent("EventOrderDeliver")    // 发货
	EventOrderReceive    = fsm.FSMEvent("EventOrderReceive")    // 收货
)

var orderFSM *fsm.FSM

// orderTransitionFunc 订单状态转移函数
func orderTransitionFunc(params map[string]interface{}, srcState fsm.FSMState, dstState fsm.FSMState) error {
	// 从params中解析order参数
	key, ok := params["order"]
	if !ok {
		return fmt.Errorf("params[\"order\"]不存在。")
	}
	curOrder := key.(*database.Order)
	fmt.Printf("order.ID: %v, order.State: %v\n", curOrder.ID, curOrder.State)

	// 订单状态转移
	if err := database.UpdateOrderState(curOrder, int(srcState), int(dstState)); err != nil {
		return err
	}
	return nil
}

// Init 状态机的状态转移图初始化
func Init() {
	orderFSM = fsm.CreateNewFSM(orderTransitionFunc)
	orderFSM.SetTransitionMap(StateOrderInit, EventOrderPlace, StateOrderToBePaid, PlaceAction{})                  // 初始化+下单 -> 待支付
	orderFSM.SetTransitionMap(StateOrderToBePaid, EventOrderPay, StateOrderToBeDelivered, PayAction{})             // 待支付+支付 -> 待发货
	orderFSM.SetTransitionMap(StateOrderToBePaid, EventOrderPayTimeout, StateOrderCancel, nil)                     // 待支付+支付超时 -> 订单取消
	orderFSM.SetTransitionMap(StateOrderToBeDelivered, EventOrderDeliver, StateOrderToBeReceived, DeliverAction{}) // 待发货+发货 -> 待收货
	orderFSM.SetTransitionMap(StateOrderToBeReceived, EventOrderReceive, StateOrderDone, ReceiveAction{})          // 待收货+收货 -> 订单完成
}

// ExecOrderTask 执行订单任务，推动状态转移
func ExecOrderTask(params map[string]interface{}) error {
	// 从params中解析order参数
	key, ok := params["order"]
	if !ok {
		return fmt.Errorf("params[\"order\"]不存在。")
	}
	curOrder := key.(*database.Order)

	// 初始化+下单 -> 待支付
	if curOrder.State == int(StateOrderInit) {
		if err := orderFSM.Push(nil, params, StateOrderInit, EventOrderPlace); err != nil {
			return err
		}
	}
	// 待支付+支付 -> 待发货
	if curOrder.State == int(StateOrderToBePaid) {
		if err := orderFSM.Push(nil, params, StateOrderToBePaid, EventOrderPay); err != nil {
			return err
		}
	}
	// 待支付+支付超时 -> 订单取消
	if curOrder.State == int(StateOrderToBePaid) {
		if err := orderFSM.Push(nil, params, StateOrderToBePaid, EventOrderPayTimeout); err != nil {
			return err
		}
	}
	// 待发货+发货 -> 待收货
	if curOrder.State == int(StateOrderToBeDelivered) {
		if err := orderFSM.Push(nil, params, StateOrderToBeDelivered, EventOrderDeliver); err != nil {
			return err
		}
	}
	// 待收货+收货 -> 订单完成
	if curOrder.State == int(StateOrderToBeReceived) {
		if err := orderFSM.Push(nil, params, StateOrderToBeReceived, EventOrderReceive); err != nil {
			return err
		}
	}
	return nil
}
