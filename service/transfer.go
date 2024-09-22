package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/dao"
	"github.com/zjn-zjn/fisher/model"
)

const (
	bagStateUniqueKey = "%d_%d"
)

// Transfer 物品转移
func Transfer(ctx context.Context, req *model.TransferReq) error {
	// check args
	err := checkArgs(req)
	if err != nil {
		return basic.NewParamsError(err)
	}
	//处理官方账户,为避免官方账户的热点问题,官方账户采用区间制
	handleOfficialBag(req)
	//创建转移事务
	state, err := dao.GetOrCreateState(ctx, req)
	if err != nil {
		return err
	}
	if state.Status == basic.StateStatusSuccess || state.Status == basic.StateStatusHalfSuccess {
		//当前事务已成功，为保证幂等性，直接返回成功
		//这里就不怕因网络问题导致的，转移在当前已经成功，但已收到回滚消息，但转移的重试进行到这里，直接返回成功导致的业务异常？
		//不怕，因为rollback一定是发生于正常执行之后，如果已经发生了rollback,那么之前的转移一定是失效的，因此返回转移成功也不会有问题
		//举例：A调用B，A调用超时，B实际已成功，A一次重试调用B，由于发生失败，A于是决定回滚
		//但由于网络问题，到了B这里，可能是重试先到，rollback后到，此时B的重试依然返回了当前转移已成功给到A，无论B的这次返回是否触达A，A都不会将转移继续进行下去
		return nil
	}

	if state.Status == basic.StateStatusRollbackDoing || state.Status == basic.StateStatusRollbackDone {
		//当前事务正在回滚或已回滚，直接返回失败
		return basic.AlreadyRolledBackErr
	}

	if state.Status != basic.StateStatusDoing {
		//当前非doing 直接报错
		return basic.StateMutationErr
	}
	deductionTxs, increaseTxs, err := itemTransferSequences(req)
	if err != nil {
		return err
	}

	err = dao.TxWrapper(ctx, state, deductionTxs, increaseTxs, req.UseHalfSuccess)
	if err != nil {
		return err
	}
	return nil
}

// 离散官方背包
func handleOfficialBag(req *model.TransferReq) {
	bagId := findFirstBagId(req)
	if bagId == nil {
		//无对标的背包ID，为官方账户之间流转，直接采用原始值
		return
	}
	musk := basic.GetRemain(*bagId)
	//修改官方账户
	for _, v := range req.FromBags {
		v := v
		if basic.IsOfficialBag(v.BagId) {
			v.BagId = basic.GetMixOfficialBagId(v.BagId, musk)
		}
	}
	for _, v := range req.ToBags {
		v := v
		if basic.IsOfficialBag(v.BagId) {
			v.BagId = basic.GetMixOfficialBagId(v.BagId, musk)
		}
	}
}

func findFirstBagId(req *model.TransferReq) *int64 {
	for _, v := range req.FromBags {
		bagId := v.BagId
		if !basic.IsOfficialBag(bagId) {
			return &bagId
		}
	}
	for _, v := range req.ToBags {
		bagId := v.BagId
		if !basic.IsOfficialBag(bagId) {
			return &bagId
		}
	}
	return nil
}

func checkArgs(req *model.TransferReq) error {
	if req.TransferId <= 0 {
		return errors.New("transferId illegal")
	}
	if req.ItemType <= 0 {
		return errors.New("item type illegal")
	}
	if req.TransferScene <= 0 {
		return errors.New("transfer scene illegal")
	}
	if len(req.FromBags) <= 0 {
		return errors.New("from bags illegal")
	}
	if len(req.ToBags) <= 0 {
		return errors.New("to bags illegal")
	}
	var fromTotalAmount int64
	var totalToAmount int64
	uniqueCheck := make(map[string]bool)
	for _, fromBagInfo := range req.FromBags {
		if fromBagInfo.BagId <= 0 {
			return fmt.Errorf("[fisher] from bag illegal from_bag:%d", fromBagInfo.BagId)
		}
		if basic.IsOfficialBag(fromBagInfo.BagId) {
			//官方背包的传入必须用枚举，由转移系统离散
			if !basic.CheckTransferOfficialBag(fromBagInfo.BagId) {
				return fmt.Errorf("[fisher] offical bag illegal from_bag:%d", fromBagInfo.BagId)
			}
		}
		//非官方背包不允许操作为0的金额
		if fromBagInfo.Amount <= 0 && !basic.IsOfficialBag(fromBagInfo.BagId) {
			return fmt.Errorf("[fisher] from amount illegal from_amount:%d", fromBagInfo.Amount)
		}
		fromTotalAmount = fromBagInfo.Amount + fromTotalAmount
		//用于校验背包转移场景是否重复，如：背包A1向A2和A3转账，A1 A2 A3的扣钱和加钱场景不应有重叠(比如A2 A3不能都是版权结算)
		fromBagUniqueKey := fmt.Sprintf(bagStateUniqueKey, fromBagInfo.BagId, fromBagInfo.ChangeType)
		if _, ok := uniqueCheck[fromBagUniqueKey]; ok {
			return fmt.Errorf("[fisher] bag duplicate change type bag:%d", fromBagInfo.BagId)
		}
		uniqueCheck[fromBagUniqueKey] = true
	}
	for _, toBagInfo := range req.ToBags {
		if toBagInfo.BagId <= 0 {
			return fmt.Errorf("[fisher] to bag illegal to_bag:%d", toBagInfo.BagId)
		}
		if basic.IsOfficialBag(toBagInfo.BagId) {
			//官方背包的传入必须用枚举，由转移系统离散
			if !basic.CheckTransferOfficialBag(toBagInfo.BagId) {
				return fmt.Errorf("[fisher] offical bag illegal to_bag:%d", toBagInfo.BagId)
			}
		}
		//非官方背包不允许操作为0的金额
		if toBagInfo.Amount <= 0 && !basic.IsOfficialBag(toBagInfo.BagId) {
			return fmt.Errorf("[fisher] to amount illegal to_amount:%d", toBagInfo.Amount)
		}
		totalToAmount = toBagInfo.Amount + totalToAmount
		//用于校验背包转移场景是否重复，如：背包A1向A2和A3转账，A1 A2 A3的扣钱和加钱场景不应有重叠(比如A2 A3不能都是版权结算)
		toBagUniqueKey := fmt.Sprintf(bagStateUniqueKey, toBagInfo.BagId, toBagInfo.ChangeType)
		if _, ok := uniqueCheck[toBagUniqueKey]; ok {
			return fmt.Errorf("[fisher] bag duplicate change type bag:%d", toBagInfo.BagId)
		}
		uniqueCheck[toBagUniqueKey] = true
	}
	if fromTotalAmount != totalToAmount {
		return errors.New("[fisher] unable to balance accounts")
	}
	return nil
}

func itemTransferSequences(req *model.TransferReq) ([]*dao.TransferTxItem, []*dao.TransferTxItem, error) {
	var fromTxs []*dao.TransferTxItem
	var toTxs []*dao.TransferTxItem

	for _, fromBagInfo := range req.FromBags {
		fromBagInfo := fromBagInfo
		fromTxs = append(fromTxs, &dao.TransferTxItem{
			Exec: func(ctx context.Context) error {
				err := dao.DeductionBag(ctx, fromBagInfo.BagId, req.TransferId, fromBagInfo.Amount, req.ItemType, req.TransferScene, basic.RecordStatusNormal, fromBagInfo.ChangeType, req.Comment)
				if err != nil {
					return err
				}
				return nil
			},
			Rollback: func(ctx context.Context) error {
				err := dao.IncreaseBag(ctx, fromBagInfo.BagId, req.TransferId, fromBagInfo.Amount, req.ItemType, req.TransferScene, basic.RecordStatusRollback, fromBagInfo.ChangeType, req.Comment)
				if err != nil {
					return err
				}
				return nil
			},
		})
	}
	for _, toBagInfo := range req.ToBags {
		toBagInfo := toBagInfo
		toTxs = append(toTxs, &dao.TransferTxItem{
			Exec: func(ctx context.Context) error {
				comment := req.Comment
				if toBagInfo.Comment != "" {
					comment = toBagInfo.Comment
				}
				err := dao.IncreaseBag(ctx, toBagInfo.BagId, req.TransferId, toBagInfo.Amount, req.ItemType, req.TransferScene, basic.RecordStatusNormal, toBagInfo.ChangeType, comment)
				if err != nil {
					return err
				}
				return nil
			},
			Rollback: func(ctx context.Context) error {
				err := dao.DeductionBag(ctx, toBagInfo.BagId, req.TransferId, toBagInfo.Amount, req.ItemType, req.TransferScene, basic.RecordStatusRollback, toBagInfo.ChangeType, req.Comment)
				if err != nil {
					return err
				}
				return nil
			},
		})
	}
	return fromTxs, toTxs, nil
}
