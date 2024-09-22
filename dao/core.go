package dao

import (
	"context"

	"gorm.io/gorm"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"
)

//==============================================================================
// 转移case
// 扣减A 10，给A 1分成 给B 8分成 给C 1分成
//
//	       A  ----------转移---------->  A                   B                     C
//
// -10(Normal&Deduct)(成功)     +1(Normal&Add)(成功)   +8(Normal&Add)(成功)    +1(Normal&Add)(失败)(操作失败本地事务回滚，无流水记录)
// --------------------准备回滚-------------------------
// +10(Rollback&Add)           -1(Rollback&Deduct)    -8(Rollback&Deduct)    -1(Rollback&Deduct)(无增的记录，空回滚)
//==============================================================================

// DeductionBag
// 扣减物品并记录转移
// 1 查询转移
// 1.1 转移存在，并状态一致，为保证幂等性，不做操作，直接返回
// 1.2 如果是正常操作，需要校验是否有同等的回滚操作已执行，如有则直接报错返回！！！！
// 1.3 如果是回滚操作，需要确认之前是否执行过加的操作，未执行过加直接结束
// 2 获取背包物品数量信息，校验物品数量
// 2.1 如果是正常操作，需要校验物品是否充足 (官方账号除外)
// 2.2 如果是回滚操作，支持将物品回滚到负数
// 3 进行扣减数量操作
func DeductionBag(ctx context.Context, bagId, transferId, amount int64, itemType basic.ItemType, transferScene basic.TransferScene, transferStatus basic.RecordStatus, changeType basic.ChangeType, comment string) error {
	transferType := getRecordTypeWithStatus(basic.RecordTypeDeduct, transferStatus)
	//背包查询和创建放在最外面，提高并发性能
	bag, err := getBagDefaultCreate(ctx, bagId, itemType)
	if err != nil {
		return err
	}
	//如果是正常操作，需要校验是否有足够的金额进行扣减，官方背包账号除外，如果是回滚操作，支持扣减到负数
	if transferStatus == basic.RecordStatusNormal && !basic.IsOfficialBag(bagId) {
		if bag.Amount < amount {
			return basic.InsufficientAmountErr
		}
	}
	originRecord, err := GetRecord(ctx, bagId, transferId, itemType, transferScene, transferType, changeType)
	if err != nil {
		return err
	}
	if originRecord != nil && originRecord.TransferStatus == transferStatus {
		//该操作已完成，直接幂等结束
		return nil
	}
	err = BagDBTX(ctx, func(ctx context.Context, db *gorm.DB) error {
		if originRecord == nil {
			if transferStatus == basic.RecordStatusRollback {
				//如果是回滚操作，需要确认之前是否执行过加的操作，未执行过加直接结束
				record := assembleRecord(transferId, bagId, amount, transferScene, basic.RecordStatusEmptyRollback, transferType, changeType, itemType, comment)
				if err = CreateRecord(ctx, &record, db); err != nil {
					return err
				}
				return nil
			}
			//写订单记录
			record := assembleRecord(transferId, bagId, amount, transferScene, transferStatus, transferType, changeType, itemType, comment)
			if err = CreateRecord(ctx, &record, db); err != nil {
				return err
			}
		}
		if transferStatus == basic.RecordStatusNormal {
			//如果是正常操作，需要校验是否有同等的回滚操作已执行，如有则直接报错返回！！！！
			if originRecord != nil && (originRecord.TransferStatus == basic.RecordStatusRollback || originRecord.TransferStatus == basic.RecordStatusEmptyRollback) {
				//之前有同等的回滚操作已执行，直接返回error
				return basic.StateMutationErr
			}
		}
		//更新订单
		if originRecord != nil {
			affect, err := UpdateRecord(ctx, bagId, transferId, itemType, transferScene, transferType, transferStatus, basic.RecordStatusNormal, changeType, db)
			if err != nil {
				return err
			}
			if !affect {
				//该操作已完成，直接结束
				return nil
			}
		}
		//订单写入/更新成功，对背包进行操作
		if amount == 0 {
			//如果操作的是0元，直接结束(一般用于某些官方账号加0操作，只记录转移不加钱)
			return nil
		}
		return deductBagAmount(ctx, bagId, amount, itemType, transferStatus, db)
	})
	if err != nil {
		return err
	}
	return nil
}

// IncreaseBag
// 增加物品并记录转移
// 1 查询转移
// 1.1 转移存在，并状态一致，为保证幂等性，不做操作，直接返回
// 1.2 如果是正常操作，需要校验是否有同等的回滚操作已执行，如有则直接报错返回！！！！
// 1.3 如果是回滚操作，需要确认之前是否执行过减的操作，未执行过减直接结束
// 2 获取背包物品数量信息
// 3 进行增加数量操作
func IncreaseBag(ctx context.Context, bagId, transferId, amount int64, itemType basic.ItemType, transferScene basic.TransferScene, transferStatus basic.RecordStatus, changeType basic.ChangeType, comment string) error {
	transferType := getRecordTypeWithStatus(basic.RecordTypeAdd, transferStatus)
	//不存在则创建放到最外面，提高并发性能
	_, err := getBagDefaultCreate(ctx, bagId, itemType)
	if err != nil {
		return err
	}
	originRecord, err := GetRecord(ctx, bagId, transferId, itemType, transferScene, transferType, changeType)
	if err != nil {
		return err
	}
	if originRecord != nil && originRecord.TransferStatus == transferStatus {
		//该操作已完成，直接幂等结束
		return nil
	}
	err = BagDBTX(ctx, func(ctx context.Context, db *gorm.DB) error {
		if originRecord == nil {
			if transferStatus == basic.RecordStatusRollback {
				//如果是回滚操作，需要确认之前是否执行过减的操作，未执行过减直接结束
				record := assembleRecord(transferId, bagId, amount, transferScene, basic.RecordStatusEmptyRollback, transferType, changeType, itemType, comment)
				if err = CreateRecord(ctx, &record, db); err != nil {
					return err
				}
				return nil
			}
			record := assembleRecord(transferId, bagId, amount, transferScene, transferStatus, transferType, changeType, itemType, comment)
			if err = CreateRecord(ctx, &record, db); err != nil {
				return err
			}
		}
		if transferStatus == basic.RecordStatusNormal {
			//如果是正常操作，需要校验是否有同等的回滚操作已执行，如有则直接报错返回！！！！
			if originRecord != nil && (originRecord.TransferStatus == basic.RecordStatusRollback || originRecord.TransferStatus == basic.RecordStatusEmptyRollback) {
				//之前有同等的回滚操作已执行，直接返回error
				return basic.StateMutationErr
			}
		}
		//更新订单
		if originRecord != nil {
			affect, err := UpdateRecord(ctx, bagId, transferId, itemType, transferScene, transferType, transferStatus, basic.RecordStatusNormal, changeType, db)
			if err != nil {
				return err
			}
			if !affect {
				//该操作已完成，返回成功
				return nil
			}
		}
		if amount == 0 {
			//如果金额是0，直接成功返回(一般用于某些官方账号加0操作，只记录转移不加钱)
			return nil
		}
		return increaseBagAmount(ctx, bagId, amount, itemType, db)
	})
	if err != nil {
		return err
	}
	return nil
}

func assembleRecord(transferId, bagId, amount int64, transferScene basic.TransferScene, transferStatus basic.RecordStatus, transferType basic.RecordType, changeType basic.ChangeType, itemType basic.ItemType, comment string) model.Record {
	return model.Record{
		TransferId:     transferId,
		TransferScene:  transferScene,
		BagId:          bagId,
		Amount:         amount,
		ItemType:       itemType,
		TransferStatus: transferStatus,
		RecordType:     transferType,
		ChangeType:     changeType,
		Comment:        comment,
	}
}
