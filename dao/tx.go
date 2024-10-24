package dao

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"
)

type TransferTxItem struct {
	Exec     func(ctx context.Context) error
	Rollback func(ctx context.Context) error
}

func ExecuteTransfer(ctx context.Context, state *model.State, deductionTxItems, increaseTxItems []*TransferTxItem, useHalfSuccess bool) error {
	if err := executeTransactions(ctx, deductionTxItems); err != nil {
		fastRollBack(ctx, state, append(increaseTxItems, deductionTxItems...))
		return err
	}

	if useHalfSuccess {
		return handleHalfSuccessTransfer(ctx, state, increaseTxItems)
	}

	if err := executeTransactions(ctx, increaseTxItems); err != nil {
		fastRollBack(ctx, state, append(increaseTxItems, deductionTxItems...))
		return err
	}

	return finalizeTransfer(ctx, state)
}

func executeTransactions(ctx context.Context, txItems []*TransferTxItem) error {
	for _, item := range txItems {
		if err := item.Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func handleHalfSuccessTransfer(ctx context.Context, state *model.State, increaseTxItems []*TransferTxItem) error {
	affected, err := UpdateStateStatusWithAffect(ctx, state.TransferId, state.TransferScene, basic.StateStatusDoing, basic.StateStatusHalfSuccess)
	if err != nil {
		return err
	}

	if !affected {
		currentState, err := GetState(ctx, state.TransferId, state.TransferScene, nil)
		if err != nil {
			return err
		}

		switch currentState.Status {
		case basic.StateStatusSuccess, basic.StateStatusHalfSuccess:
			return nil
		case basic.StateStatusRollbackDone:
			return basic.StateMutationErr
		default:
			fastRollBack(ctx, state, increaseTxItems)
			return basic.StateMutationErr
		}
	}

	go func() {
		if err := executeTransactions(ctx, increaseTxItems); err != nil {
			return
		}
		_, _ = UpdateStateStatusWithAffect(ctx, state.TransferId, state.TransferScene, basic.StateStatusHalfSuccess, basic.StateStatusSuccess)
	}()

	return nil
}

func finalizeTransfer(ctx context.Context, state *model.State) error {
	affected, err := UpdateStateStatusWithAffect(ctx, state.TransferId, state.TransferScene, basic.StateStatusDoing, basic.StateStatusSuccess)
	if err != nil {
		return err
	}

	if !affected {
		currentState, err := GetState(ctx, state.TransferId, state.TransferScene, nil)
		if err != nil {
			return err
		}

		switch currentState.Status {
		case basic.StateStatusSuccess, basic.StateStatusHalfSuccess:
			return nil
		case basic.StateStatusRollbackDone:
			return basic.StateMutationErr
		default:
			return basic.StateMutationErr
		}
	}

	return nil
}

func fastRollBack(ctx context.Context, state *model.State, txItems []*TransferTxItem) {
	affected, err := UpdateStateStatusWithAffect(ctx, state.TransferId, state.TransferScene, basic.StateStatusDoing, basic.StateStatusRollbackDoing)
	if err != nil || !affected {
		return
	}

	for _, tx := range txItems {
		if err := tx.Rollback(ctx); err != nil {
			return
		}
	}

	_ = UpdateStateStatus(ctx, state.TransferId, state.TransferScene, basic.StateStatusRollbackDoing, basic.StateStatusRollbackDone)
}

func RecordAndBagInstanceTX(ctx context.Context, bagId int64, fn func(context.Context, *gorm.DB) error) error {
	return executeTx(basic.GetRecordAndBagWriteDB(ctx, bagId), fn)
}

func StateInstanceTX(ctx context.Context, transferId int64, fn func(context.Context, *gorm.DB) error) error {
	return executeTx(basic.GetStateWriteDB(ctx, transferId), fn)
}

func executeTx(db *gorm.DB, fn func(context.Context, *gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return basic.NewWithErr(basic.DBFailedErrCode, errors.Wrap(tx.Error, "[fisher] begin tx failed"))
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := fn(tx.Statement.Context, tx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return basic.NewWithErr(basic.DBFailedErrCode, errors.Wrap(rbErr, "[fisher] rollback tx failed"))
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return basic.NewWithErr(basic.DBFailedErrCode, errors.Wrap(err, "[fisher] commit tx failed"))
	}

	return nil
}
