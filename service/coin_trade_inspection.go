package service

import (
	"context"
	"github.com/zjn-zjn/coin-trade/basic"
	"github.com/zjn-zjn/coin-trade/dao"
	"github.com/zjn-zjn/coin-trade/model"
)

// CoinTradeInspection 拿到截止lastTime还在进行中(doing、rollback doing 和 half success)的交易，进行推进
func CoinTradeInspection(ctx context.Context, lastTime int64) []error {
	//获取需要推进的交易
	stateList, err := dao.GetNeedInspectionTradeStateList(ctx, lastTime)
	if err != nil {
		return []error{err}
	}
	if len(stateList) == 0 {
		return nil
	}
	//推进交易
	var errs []error
	for _, state := range stateList {
		state := state
		if state.Status == basic.TradeStateStatusHalfSuccess {
			//推进成功
			err = processHalfSuccessState(ctx, state)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			continue
		}
		//不存在需要推进doing的情况，doing没有变成success就是失败了
		if state.Status == basic.TradeStateStatusRollbackDoing || state.Status == basic.TradeStateStatusDoing {
			//推进回滚
			err = RollbackTrade(ctx, &model.RollbackTradeReq{
				TradeId:    state.TradeId,
				TradeScene: state.TradeScene,
			})
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
	}
	return errs
}

// processHalfSuccessState HalfSuccess的推进应该极力保证成功,所以没有回滚操作
func processHalfSuccessState(ctx context.Context, state *model.TradeState) error {
	txs, err := tradeProcessHalfSuccessCoinTxSequences(state)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		err = tx.Exec(ctx)
		if err != nil {
			return err
		}
	}
	err = dao.UpdateTradeStateStatus(ctx, state.TradeId, state.TradeScene, basic.TradeStateStatusHalfSuccess, basic.TradeStateStatusSuccess)
	return err
}

// HalfSuccess的推进应该极力保证成功,所以没有回滚操作
func tradeProcessHalfSuccessCoinTxSequences(state *model.TradeState) ([]dao.TradeTxItem, error) {
	//扣除金额一定是已经成功，所以这里不会再有扣除动作
	//增加金额
	var txs = make([]dao.TradeTxItem, 0)
	for _, toWalletInfo := range state.ToWallets {
		txs = append(txs, dao.TradeTxItem{
			Exec: func(ctx context.Context) error {
				// 增加金额
				comment := state.Comment
				if toWalletInfo.Comment != "" {
					comment = toWalletInfo.Comment
				}
				err := dao.IncreaseWallet(ctx, toWalletInfo.WalletId, state.TradeId, toWalletInfo.Amount, state.CoinType, state.TradeScene, basic.TradeRecordStatusNormal, toWalletInfo.ChangeType, comment)
				if err != nil {
					return err
				}
				return nil
			},
		})
	}
	return txs, nil
}
