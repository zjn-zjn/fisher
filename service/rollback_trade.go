package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/zjn-zjn/coin-trade/dao"

	"gorm.io/gorm"

	"github.com/zjn-zjn/coin-trade/basic"
	"github.com/zjn-zjn/coin-trade/model"
)

// RollbackTrade 回滚交易
// 由于上游业务问题，交易后的动作无法继续推行下去，于是调用回滚对之前已成功的交易进行回滚
// 1. 当前已回滚成功，直接返回
// 2. 将非回滚成功状态转化成回滚中，执行回滚操作
func RollbackTrade(ctx context.Context, req *model.RollbackTradeReq) error {
	if req == nil || req.TradeId == 0 || req.TradeScene == 0 {
		return basic.NewParamsError(errors.New("[coin-trade] rollback trade params error"))
	}
	var state *model.TradeState
	err := dao.WalletDBTX(ctx, func(ctx context.Context, db *gorm.DB) error {
		var err error
		state, err = dao.GetTradeState(ctx, req.TradeId, req.TradeScene, db)
		if err != nil {
			return err
		}
		if state == nil {
			state = model.AssembleTradeState(nil, 0, req.TradeId,
				0, req.TradeScene, false, basic.TradeStateStatusRollbackDone, 0, "empty rollback")
			//不存在, 有可能正在写入中，可能属于回滚早到的情况，记录一条空回滚成功，避免后到的交易正常执行
			//举例 A调用B 超时，A触发回滚  由于网络问题，回滚先行到达B，交易后到达B
			//如果回滚成功，不做记录，A认为回滚成功，那么交易到达B时可能会触发正常交易
			//那为什么找不到事务就直接返回报错呢？
			//因为如果A调用B的交易，不能正常触达B，B无法生成交易记录，那么这次回滚就一直不能够成功
			if err = dao.CreateTradeState(ctx, state, db); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if state.Status == basic.TradeStateStatusRollbackDone {
		//已成功回滚，直接return
		return nil
	}
	if state.Status != basic.TradeStateStatusRollbackDoing {
		affect, err := dao.UpdateTradeStateToRollbackDoing(ctx, req.TradeId, req.TradeScene)
		if err != nil {
			return err
		}
		if !affect {
			//无更新，说明已经有其他的回滚在进行中或已结束
			return nil
		}
	}
	//这里都是需要回滚的情况
	if state.Inverse {
		//对增的钱包进行扣减
		err = dao.DeductionWallet(ctx, state.FromWalletId, state.TradeId, state.FromAmount, state.CoinType, req.TradeScene, basic.TradeRecordStatusRollback, basic.ChangeType(state.TradeScene), fmt.Sprintf("rollback %s", state.Comment))
		if err != nil {
			return err
		}
		//对减的钱包进行增
		for _, v := range state.ToWallets {
			v := v
			err = dao.IncreaseWallet(ctx, v.WalletId, state.TradeId, v.Amount, state.CoinType, req.TradeScene, basic.TradeRecordStatusRollback, basic.ChangeType(v.AddType), fmt.Sprintf("rollback %s", v.Comment))
			if err != nil {
				return err
			}
		}
		err = dao.UpdateTradeStateStatus(ctx, req.TradeId, req.TradeScene, basic.TradeStateStatusRollbackDoing, basic.TradeStateStatusRollbackDone)
		if err != nil {
			return err
		}
		return nil
	}
	//对加的钱包进行扣减
	for _, v := range state.ToWallets {
		v := v
		err = dao.DeductionWallet(ctx, v.WalletId, state.TradeId, v.Amount, state.CoinType, req.TradeScene, basic.TradeRecordStatusRollback, basic.ChangeType(v.AddType), fmt.Sprintf("rollback %s", v.Comment))
		if err != nil {
			return err
		}
	}
	//对扣减的钱包进行增
	err = dao.IncreaseWallet(ctx, state.FromWalletId, state.TradeId, state.FromAmount, state.CoinType, req.TradeScene, basic.TradeRecordStatusRollback, basic.ChangeType(state.TradeScene), fmt.Sprintf("rollback %s", state.Comment))
	if err != nil {
		return err
	}
	err = dao.UpdateTradeStateStatus(ctx, req.TradeId, req.TradeScene, basic.TradeStateStatusRollbackDoing, basic.TradeStateStatusRollbackDone)
	if err != nil {
		return err
	}
	return nil
}
