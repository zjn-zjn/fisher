package service

import (
	"context"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/dao"
	"github.com/zjn-zjn/fisher/model"
)

// Inspection 拿到截止lastTime还在进行中(doing、rollback doing 和 half success)的转移，进行推进
func Inspection(ctx context.Context, lastTime int64) []error {
	//获取需要推进的转移
	stateList, err := dao.GetNeedInspectionStateList(ctx, lastTime)
	if err != nil {
		return []error{err}
	}
	if len(stateList) == 0 {
		return nil
	}
	//推进转移
	var errs []error
	for _, state := range stateList {
		state := state
		if state.Status == basic.StateStatusHalfSuccess {
			//推进成功
			err = processHalfSuccessState(ctx, state)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			continue
		}
		//不存在需要推进doing的情况，doing没有变成success就是失败了
		if state.Status == basic.StateStatusRollbackDoing || state.Status == basic.StateStatusDoing {
			//推进回滚
			err = Rollback(ctx, &model.RollbackReq{
				TransferId:    state.TransferId,
				TransferScene: state.TransferScene,
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
func processHalfSuccessState(ctx context.Context, state *model.State) error {
	txs, err := processHalfSuccessTxSequences(state)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		err = tx.Exec(ctx)
		if err != nil {
			return err
		}
	}
	err = dao.UpdateStateStatus(ctx, state.TransferId, state.TransferScene, basic.StateStatusHalfSuccess, basic.StateStatusSuccess)
	return err
}

// HalfSuccess的推进应该极力保证成功,所以没有回滚操作
func processHalfSuccessTxSequences(state *model.State) ([]dao.TransferTxItem, error) {
	//扣除金额一定是已经成功，所以这里不会再有扣除动作
	//增加金额
	var txs = make([]dao.TransferTxItem, 0)
	for _, toAccountInfo := range state.ToAccounts {
		txs = append(txs, dao.TransferTxItem{
			Exec: func(ctx context.Context) error {
				// 增加金额
				comment := state.Comment
				if toAccountInfo.Comment != "" {
					comment = toAccountInfo.Comment
				}
				err := dao.IncreaseAccount(ctx, toAccountInfo.AccountId, state.TransferId, toAccountInfo.Amount, toAccountInfo.ItemType, state.TransferScene, basic.RecordStatusNormal, toAccountInfo.ChangeType, comment)
				if err != nil {
					return err
				}
				return nil
			},
		})
	}
	return txs, nil
}
