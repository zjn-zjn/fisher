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

	handleOfficialBags(req)

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
	if req.TransferId <= 0 || req.ItemType <= 0 || req.TransferScene <= 0 {
		return errors.New("invalid transfer parameters")
	}

	if len(req.FromBags) == 0 || len(req.ToBags) == 0 {
		return errors.New("empty from or to bags")
	}

	uniqueBags := make(map[string]struct{})
	var fromTotal, toTotal int64

	for _, bag := range req.FromBags {
		if err := validateBag(bag, "from", uniqueBags); err != nil {
			return err
		}
		fromTotal += bag.Amount
	}

	for _, bag := range req.ToBags {
		if err := validateBag(bag, "to", uniqueBags); err != nil {
			return err
		}
		toTotal += bag.Amount
	}

	if fromTotal != toTotal {
		return errors.New("unbalanced transfer amounts")
	}

	return nil
}

func validateBag(bag *model.TransferItem, bagType string, uniqueBags map[string]struct{}) error {
	if bag.BagId <= 0 {
		return fmt.Errorf("invalid %s bag ID: %d", bagType, bag.BagId)
	}

	if basic.IsOfficialBag(bag.BagId) {
		if !basic.CheckTransferOfficialBag(bag.BagId) {
			return fmt.Errorf("invalid official %s bag: %d", bagType, bag.BagId)
		}
	} else if bag.Amount <= 0 {
		return fmt.Errorf("invalid %s amount: %d", bagType, bag.Amount)
	}

	uniqueKey := fmt.Sprintf("%d_%d", bag.BagId, bag.ChangeType)
	if _, exists := uniqueBags[uniqueKey]; exists {
		return fmt.Errorf("duplicate change type for bag: %d", bag.BagId)
	}
	uniqueBags[uniqueKey] = struct{}{}

	return nil
}

func handleOfficialBags(req *model.TransferReq) {
	musk := findFirstNonOfficialBagMusk(req)
	if musk == nil {
		return
	}

	updateBagIds := func(bags []*model.TransferItem) {
		for i := range bags {
			if basic.IsOfficialBag(bags[i].BagId) {
				bags[i].BagId = basic.GetMixOfficialBagId(bags[i].BagId, *musk)
			}
		}
	}

	updateBagIds(req.FromBags)
	updateBagIds(req.ToBags)
}

func findFirstNonOfficialBagMusk(req *model.TransferReq) *int64 {
	findMusk := func(bags []*model.TransferItem) *int64 {
		for _, bag := range bags {
			if !basic.IsOfficialBag(bag.BagId) {
				musk := basic.GetRemain(bag.BagId)
				return &musk
			}
		}
		return nil
	}

	if musk := findMusk(req.FromBags); musk != nil {
		return musk
	}
	return findMusk(req.ToBags)
}

func prepareTransferTransactions(req *model.TransferReq) ([]*dao.TransferTxItem, []*dao.TransferTxItem, error) {
	fromTxs := make([]*dao.TransferTxItem, 0, len(req.FromBags))
	toTxs := make([]*dao.TransferTxItem, 0, len(req.ToBags))

	for _, fromBag := range req.FromBags {
		fromTxs = append(fromTxs, createDeductionTx(req, fromBag))
	}

	for _, toBag := range req.ToBags {
		toTxs = append(toTxs, createIncreaseTx(req, toBag))
	}

	return fromTxs, toTxs, nil
}

func createDeductionTx(req *model.TransferReq, bag *model.TransferItem) *dao.TransferTxItem {
	return &dao.TransferTxItem{
		Exec: func(ctx context.Context) error {
			return dao.DeductionBag(ctx, bag.BagId, req.TransferId, bag.Amount, req.ItemType, req.TransferScene, basic.RecordStatusNormal, bag.ChangeType, req.Comment)
		},
		Rollback: func(ctx context.Context) error {
			return dao.IncreaseBag(ctx, bag.BagId, req.TransferId, bag.Amount, req.ItemType, req.TransferScene, basic.RecordStatusRollback, bag.ChangeType, req.Comment)
		},
	}
}

func createIncreaseTx(req *model.TransferReq, bag *model.TransferItem) *dao.TransferTxItem {
	return &dao.TransferTxItem{
		Exec: func(ctx context.Context) error {
			comment := req.Comment
			if bag.Comment != "" {
				comment = bag.Comment
			}
			return dao.IncreaseBag(ctx, bag.BagId, req.TransferId, bag.Amount, req.ItemType, req.TransferScene, basic.RecordStatusNormal, bag.ChangeType, comment)
		},
		Rollback: func(ctx context.Context) error {
			return dao.DeductionBag(ctx, bag.BagId, req.TransferId, bag.Amount, req.ItemType, req.TransferScene, basic.RecordStatusRollback, bag.ChangeType, req.Comment)
		},
	}
}
