package dao

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/zjn-zjn/coin-trade/model"

	"github.com/zjn-zjn/coin-trade/basic"
)

type TradeTxItem struct {
	Exec     func(ctx context.Context) error
	Rollback func(ctx context.Context) error
}

func TxWrapper(ctx context.Context, state *model.TradeState, deductionTxItem *TradeTxItem, increaseTxItems []*TradeTxItem, useHalfSuccess bool) error {

	err := deductionTxItem.Exec(ctx)
	if err != nil {
		fastRollBack(ctx, state, append(increaseTxItems, deductionTxItem))
		return err
	}
	if useHalfSuccess {
		//由Doing到HalfSuccess
		affect, err := UpdateTradeStateStatusWithAffect(ctx, state.TradeId, state.TradeScene, basic.TradeStateStatusDoing, basic.TradeStateStatusHalfSuccess)
		if err != nil {
			fastRollBack(ctx, state, append(increaseTxItems, deductionTxItem))
			return err
		}
		if !affect {
			currentState, err := GetTradeState(ctx, state.TradeId, state.TradeScene, nil)
			if err != nil {
				fastRollBack(ctx, state, append(increaseTxItems, deductionTxItem))
				return err
			}
			if currentState.Status == basic.TradeStateStatusSuccess || currentState.Status == basic.TradeStateStatusHalfSuccess {
				//已经成功了，直接幂等结束
				return nil
			}
			if currentState.Status == basic.TradeStateStatusRollbackDone {
				return basic.StateMutationErr
			}
			fastRollBack(ctx, state, append(increaseTxItems, deductionTxItem))
			return basic.StateMutationErr
		}
		go func() {
			for _, item := range increaseTxItems {
				err := item.Exec(ctx)
				if err != nil {
					return
				}
			}
			_, _ = UpdateTradeStateStatusWithAffect(ctx, state.TradeId, state.TradeScene, basic.TradeStateStatusHalfSuccess, basic.TradeStateStatusSuccess)
		}()
		return nil
	}

	for _, item := range increaseTxItems {
		err = item.Exec(ctx)
		if err != nil {
			fastRollBack(ctx, state, append(increaseTxItems, deductionTxItem))
			return err
		}
	}
	//由Doing到Success
	//如果更新是0,有两种情况(这里再根据查询的结果去判断是否需要返回成功，而不是直接返回失败，有助于提高高并发场景下的交易成功率)
	//1. 已经成功了，直接幂等返回(success并不一定是最终态，但本次直接返回成功没有问题，如果上游需要回滚，一定不会在意这次成功的返回)
	//2. 已经失败了，但已回滚完成，直接err返回
	//3. 已经失败了，但未回滚完成，err返回并推进回滚
	affect, err := UpdateTradeStateStatusWithAffect(ctx, state.TradeId, state.TradeScene, basic.TradeStateStatusDoing, basic.TradeStateStatusSuccess)
	if err != nil {
		//这个时候不知道到底是什么状态，直接按照最坏结果处理
		fastRollBack(ctx, state, append(increaseTxItems, deductionTxItem))
		return err
	}
	if !affect {
		currentState, err := GetTradeState(ctx, state.TradeId, state.TradeScene, nil)
		if err != nil {
			//这个时候不知道到底是什么状态，直接按照最坏结果处理
			fastRollBack(ctx, state, append(increaseTxItems, deductionTxItem))
			return err
		}
		//success并不一定是最终态，但本次直接返回成功没有问题
		if currentState.Status == basic.TradeStateStatusSuccess || currentState.Status == basic.TradeStateStatusHalfSuccess {
			//已经成功了，直接幂等结束
			return nil
		}
		//rollback done一定是最终态
		if currentState.Status == basic.TradeStateStatusRollbackDone {
			return basic.StateMutationErr
		}
		fastRollBack(ctx, state, append(increaseTxItems, deductionTxItem))
		return basic.StateMutationErr
	}
	return nil
}

// WalletDBTX 事务
func WalletDBTX(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	tx := basic.GetCoinTradeWriteDB(ctx).Begin()
	if tx.Error != nil {
		return basic.NewWithErr(basic.DBFailedErrCode, errors.Wrap(tx.Error, "[coin-trade] begin tx failed"))
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()
	if err := fn(ctx, tx); err != nil {
		if err := tx.Rollback().Error; err != nil {
			return basic.NewWithErr(basic.DBFailedErrCode, errors.Wrap(err, "[coin-trade] rollback tx failed"))
		}
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return basic.NewWithErr(basic.DBFailedErrCode, errors.Wrap(err, "[coin-trade] commit tx failed"))
	}
	return nil
}

// fastRollBack 为什么不需要err返回？ 需要fastRollBack的情况一定是主业务出错了，主业务一定会返回err，上游业务会再次调用回滚，自己本身也会重试回滚
func fastRollBack(ctx context.Context, state *model.TradeState, singles []*TradeTxItem) {
	affect, err := UpdateTradeStateStatusWithAffect(ctx, state.TradeId, state.TradeScene, basic.TradeStateStatusDoing, basic.TradeStateStatusRollbackDoing)
	if err != nil {
		return
	}
	if !affect {
		//已经有回滚中的操作，直接返回，如果回滚中的操作失败，会有重试推进回滚
		return
	}
	for _, tx := range singles {
		err = tx.Rollback(ctx)
		if err != nil {
			return
		}
	}
	err = UpdateTradeStateStatus(ctx, state.TradeId, state.TradeScene, basic.TradeStateStatusRollbackDoing, basic.TradeStateStatusRollbackDone)
	if err != nil {
		return
	}
}
