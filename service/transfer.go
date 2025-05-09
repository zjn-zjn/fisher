package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/dao"
	"github.com/zjn-zjn/fisher/model"
)

// Transfer 物品转移
func Transfer(ctx context.Context, req *model.TransferReq) error {
	if err := validateTransferRequest(req); err != nil {
		return basic.NewParamsError(err)
	}

	handleOfficialAccounts(req)

	state, err := dao.GetOrCreateState(ctx, req)
	if err != nil {
		return err
	}

	switch state.Status {
	case basic.StateStatusSuccess, basic.StateStatusHalfSuccess:
		return nil // 幂等处理
	case basic.StateStatusRollbackDoing, basic.StateStatusRollbackDone:
		return basic.AlreadyRolledBackErr
	case basic.StateStatusDoing:
		// 继续处理
	default:
		return basic.StateMutationErr
	}

	deductionTxs, increaseTxs, err := prepareTransferTransactions(req)
	if err != nil {
		return err
	}

	return dao.ExecuteTransfer(ctx, state, deductionTxs, increaseTxs, req.UseHalfSuccess)
}

func validateTransferRequest(req *model.TransferReq) error {
	if req.TransferId <= 0 || req.TransferScene <= 0 {
		return errors.New("invalid transfer parameters")
	}

	if len(req.FromAccounts) == 0 || len(req.ToAccounts) == 0 {
		return errors.New("empty from or to accounts")
	}

	uniqueAccounts := make(map[string]struct{})
	fromTotalMap := make(map[basic.ItemType]int64)
	toTotalMap := make(map[basic.ItemType]int64)

	for _, account := range req.FromAccounts {
		if err := validateAccount(account, "from", uniqueAccounts); err != nil {
			return err
		}
		fromTotal := fromTotalMap[account.ItemType]
		fromTotalMap[account.ItemType] = fromTotal + account.Amount
	}

	for _, account := range req.ToAccounts {
		if err := validateAccount(account, "to", uniqueAccounts); err != nil {
			return err
		}
		toTotal := toTotalMap[account.ItemType]
		toTotalMap[account.ItemType] = toTotal + account.Amount
	}

	if len(fromTotalMap) != len(toTotalMap) {
		return errors.New("unbalanced transfer amounts")
	}

	for itemType, from := range fromTotalMap {
		to := toTotalMap[itemType]
		if from != to {
			return fmt.Errorf("unbalanced transfer amounts item type:%d from:%v to:%v", itemType, from, to)
		}
	}
	return nil
}

func validateAccount(account *model.TransferItem, accountType string, uniqueAccounts map[string]struct{}) error {
	if account.AccountId <= 0 {
		return fmt.Errorf("invalid %s account ID: %d", accountType, account.AccountId)
	}

	if basic.IsOfficialAccount(account.AccountId) {
		if !basic.CheckTransferOfficialAccount(account.AccountId) {
			return fmt.Errorf("invalid official %s account: %d", accountType, account.AccountId)
		}
	} else if account.Amount <= 0 {
		return fmt.Errorf("invalid %s amount: %d", accountType, account.Amount)
	}

	uniqueKey := fmt.Sprintf("%d_%d", account.AccountId, account.ChangeType)
	if _, exists := uniqueAccounts[uniqueKey]; exists {
		return fmt.Errorf("duplicate change type for account: %d", account.AccountId)
	}
	uniqueAccounts[uniqueKey] = struct{}{}

	return nil
}

func handleOfficialAccounts(req *model.TransferReq) {
	musk := findFirstNonOfficialAccountMusk(req)
	if musk == nil {
		return
	}

	updateAccountIds := func(accounts []*model.TransferItem) {
		for i := range accounts {
			if basic.IsOfficialAccount(accounts[i].AccountId) {
				accounts[i].AccountId = basic.GetMixOfficialAccountId(accounts[i].AccountId, *musk)
			}
		}
	}

	updateAccountIds(req.FromAccounts)
	updateAccountIds(req.ToAccounts)
}

func findFirstNonOfficialAccountMusk(req *model.TransferReq) *int64 {
	findMusk := func(accounts []*model.TransferItem) *int64 {
		for _, account := range accounts {
			if !basic.IsOfficialAccount(account.AccountId) {
				musk := basic.GetRemain(account.AccountId)
				return &musk
			}
		}
		return nil
	}

	if musk := findMusk(req.FromAccounts); musk != nil {
		return musk
	}
	return findMusk(req.ToAccounts)
}

func prepareTransferTransactions(req *model.TransferReq) ([]*dao.TransferTxItem, []*dao.TransferTxItem, error) {
	fromTxs := make([]*dao.TransferTxItem, 0, len(req.FromAccounts))
	toTxs := make([]*dao.TransferTxItem, 0, len(req.ToAccounts))

	for _, fromAccount := range req.FromAccounts {
		fromTxs = append(fromTxs, createDeductionTx(req, fromAccount))
	}

	for _, toAccount := range req.ToAccounts {
		toTxs = append(toTxs, createIncreaseTx(req, toAccount))
	}

	return fromTxs, toTxs, nil
}

func createDeductionTx(req *model.TransferReq, account *model.TransferItem) *dao.TransferTxItem {
	return &dao.TransferTxItem{
		Exec: func(ctx context.Context) error {
			return dao.DeductionAccount(ctx, account.AccountId, req.TransferId, account.Amount, account.ItemType, req.TransferScene, basic.RecordStatusNormal, account.ChangeType, req.Comment)
		},
		Rollback: func(ctx context.Context) error {
			return dao.IncreaseAccount(ctx, account.AccountId, req.TransferId, account.Amount, account.ItemType, req.TransferScene, basic.RecordStatusRollback, account.ChangeType, req.Comment)
		},
	}
}

func createIncreaseTx(req *model.TransferReq, account *model.TransferItem) *dao.TransferTxItem {
	return &dao.TransferTxItem{
		Exec: func(ctx context.Context) error {
			comment := req.Comment
			if account.Comment != "" {
				comment = account.Comment
			}
			return dao.IncreaseAccount(ctx, account.AccountId, req.TransferId, account.Amount, account.ItemType, req.TransferScene, basic.RecordStatusNormal, account.ChangeType, comment)
		},
		Rollback: func(ctx context.Context) error {
			return dao.DeductionAccount(ctx, account.AccountId, req.TransferId, account.Amount, account.ItemType, req.TransferScene, basic.RecordStatusRollback, account.ChangeType, req.Comment)
		},
	}
}
