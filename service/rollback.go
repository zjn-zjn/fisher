package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/zjn-zjn/fisher/dao"

	"gorm.io/gorm"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"
)

// Rollback 回滚转移
// 由于上游业务问题，转移后的动作无法继续推行下去，于是调用回滚对之前已成功的转移进行回滚
// 1. 当前已回滚成功，直接返回
// 2. 将非回滚成功状态转化成回滚中，执行回滚操作
func Rollback(ctx context.Context, req *model.RollbackReq) error {
	if req == nil || req.TransferId == 0 || req.TransferScene == 0 {
		return basic.NewParamsError(errors.New("[fisher] rollback transfer params error"))
	}
	var state *model.State
	err := dao.StateInstanceTX(ctx, req.TransferId, func(ctx context.Context, db *gorm.DB) error {
		var err error
		state, err = dao.GetState(ctx, req.TransferId, req.TransferScene, db)
		if err != nil {
			return err
		}
		if state == nil {
			state = model.AssembleState(nil, nil, req.TransferId, req.TransferScene, basic.StateStatusRollbackDone, 0, "empty rollback")
			//不存在, 有可能正在写入中，可能属于回滚早到的情况，记录一条空回滚成功，避免后到的转移正常执行
			//举例 A调用B 超时，A触发回滚  由于网络问题，回滚先行到达B，转移后到达B
			//如果回滚成功，不做记录，A认为回滚成功，那么转移到达B时可能会触发正常转移
			//那为什么找不到事务就直接返回报错呢？
			//因为如果A调用B的转移，不能正常触达B，B无法生成转移记录，那么这次回滚就一直不能够成功
			if err = dao.CreateState(ctx, state, db); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if state.Status == basic.StateStatusRollbackDone {
		//已成功回滚，直接return
		return nil
	}
	if state.Status != basic.StateStatusRollbackDoing {
		affect, err := dao.UpdateStateToRollbackDoing(ctx, req.TransferId, req.TransferScene)
		if err != nil {
			return err
		}
		if !affect {
			//无更新，说明已经有其他的回滚在进行中或已结束
			return nil
		}
	}
	//对加的账户进行扣减
	for _, v := range state.ToAccounts {
		v := v
		err = dao.DeductionAccount(ctx, v.AccountId, state.TransferId, v.Amount, state.ItemType, req.TransferScene, basic.RecordStatusRollback, v.ChangeType, fmt.Sprintf("rollback %s", v.Comment))
		if err != nil {
			return err
		}
	}
	//对加的账户进行扣减
	for _, v := range state.FromAccounts {
		v := v
		err = dao.IncreaseAccount(ctx, v.AccountId, state.TransferId, v.Amount, state.ItemType, req.TransferScene, basic.RecordStatusRollback, v.ChangeType, fmt.Sprintf("rollback %s", v.Comment))
		if err != nil {
			return err
		}
	}
	err = dao.UpdateStateStatus(ctx, req.TransferId, req.TransferScene, basic.StateStatusRollbackDoing, basic.StateStatusRollbackDone)
	if err != nil {
		return err
	}
	return nil
}
