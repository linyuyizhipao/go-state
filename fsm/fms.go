package fsm

import (
	"fmt"
	"reflect"
	"slp/test/state/database"
)

// FSMState 状态机的状态类型
type FSMState int

// FSMEvent 状态机的事件类型
type FSMEvent string

// FSMTransitionMap 状态机的状态转移图类型，现态和事件一旦确定，次态和动作就唯一确定
type FSMTransitionMap map[FSMState]map[FSMEvent]FSMDstStateAndAction

// FSMTransitionFunc 状态机的状态转移函数类型
type FSMTransitionFunc func(params map[string]interface{}, srcState FSMState, dstState FSMState) error

// FSMDstStateAndAction 状态机的次态和动作
type FSMDstStateAndAction struct {
	DstState FSMState  // 次态
	Action   FSMAction // 动作
}

// FSMAction 状态机的动作
type FSMAction interface {
	Before(bizParams map[string]interface{}) error                   // 状态转移前执行
	Execute(bizParams map[string]interface{}, tx *database.DB) error // 状态转移中执行
	After(bizParams map[string]interface{}) error                    // 状态转移后执行
}

// FSM 状态机，元素均为不可导出
type FSM struct {
	transitionMap  FSMTransitionMap  // 状态转移图
	transitionFunc FSMTransitionFunc // 状态转移函数
}

// CreateNewFSM 创建一个新的状态机
func CreateNewFSM(transitionFunc FSMTransitionFunc) *FSM {
	return &FSM{
		transitionMap:  make(FSMTransitionMap),
		transitionFunc: transitionFunc,
	}
}

// SetTransitionMap 设置状态机的状态转移图
func (fsm *FSM) SetTransitionMap(srcState FSMState, event FSMEvent, dstState FSMState, action FSMAction) {
	if int(srcState) < 0 || len(event) <= 0 || int(dstState) < 0 {
		panic("现态|事件|次态非法。")
		return
	}
	transitionMap := fsm.transitionMap
	if transitionMap == nil {
		transitionMap = make(FSMTransitionMap)
	}
	if _, ok := transitionMap[srcState]; !ok {
		transitionMap[srcState] = make(map[FSMEvent]FSMDstStateAndAction)
	}
	if _, ok := transitionMap[srcState][event]; !ok {
		dstStateAndAction := FSMDstStateAndAction{
			DstState: dstState,
			Action:   action,
		}
		transitionMap[srcState][event] = dstStateAndAction
	} else {
		fmt.Printf("现态[%v]+事件[%v]+次态[%v]已定义过，请勿重复定义。\n", srcState, event, dstState)
		return
	}
	fsm.transitionMap = transitionMap
}

// Push 状态机的状态迁移
func (fsm *FSM) Push(tx *database.DB, params map[string]interface{}, currentState FSMState, event FSMEvent) error {
	// 根据现态和事件从状态转移图获取次态和动作
	transitionMap := fsm.transitionMap
	events, eventExist := transitionMap[currentState]
	if !eventExist {
		return fmt.Errorf("现态[%v]未配置迁移事件", currentState)
	}
	dstStateAndAction, ok := events[event]
	if !ok {
		return fmt.Errorf("现态[%v]+迁移事件[%v]未配置次态", currentState, event)
	}
	dstState := dstStateAndAction.DstState
	action := dstStateAndAction.Action
	// 执行before方法
	if action != nil {
		fsmActionName := reflect.ValueOf(action).String()
		fmt.Printf("现态[%v]+迁移事件[%v]->次态[%v], [%v].before\n", currentState, event, dstState, fsmActionName)
		if err := action.Before(params); err != nil {
			return fmt.Errorf("现态[%v]+迁移事件[%v]->次态[%v]失败, [%v].before, err: %v", currentState, event, dstState, fsmActionName, err)
		}
	}

	// 事务执行execute方法和transitionFunc
	if tx == nil {
		tx = new(database.DB)
	}
	transactionErr := tx.Transaction(func() error {
		fsmActionName := reflect.ValueOf(action).String()
		fmt.Printf("现态[%v]+迁移事件[%v]->次态[%v], [%v].execute\n", currentState, event, dstState, fsmActionName)
		if action != nil {
			if err := action.Execute(params, tx); err != nil {
				return fmt.Errorf("状态转移执行出错：%v", err)
			}
		}

		fmt.Printf("现态[%v]+迁移事件[%v]->次态[%v], transitionFunc\n", currentState, event, dstState)
		if err := fsm.transitionFunc(params, currentState, dstState); err != nil {
			return fmt.Errorf("执行状态转移函数出错: %v", err)
		}
		return nil
	})
	if transactionErr != nil {
		return transactionErr
	}

	// 执行after方法
	if action != nil {
		fsmActionName := reflect.ValueOf(action).String()
		fmt.Printf("现态[%v]+迁移事件[%v]->次态[%v], [%v].after\n", currentState, event, dstState, fsmActionName)
		if err := action.After(params); err != nil {
			return fmt.Errorf("现态[%v]+迁移事件[%v]->次态[%v]失败, [%v].before, err: %v", currentState, event, dstState, fsmActionName, err)
		}
	}
	return nil
}
