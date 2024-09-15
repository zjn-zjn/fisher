package dao

import (
	"context"
	"gorm.io/gorm"

	"github.com/zjn-zjn/coin-trade/basic"
	"github.com/zjn-zjn/coin-trade/model"
)

//==============================================================================
// 交易case
// 扣减A 10，给A 1分成 给B 8分成 给C 1分成
//
//	       A  ----------交易---------->  A                   B                     C
//
// -10(Normal&Deduct)(成功)     +1(Normal&Add)(成功)   +8(Normal&Add)(成功)    +1(Normal&Add)(失败)(操作失败本地事务回滚，无流水记录)
// --------------------准备回滚-------------------------
// +10(Rollback&Add)           -1(Rollback&Deduct)    -8(Rollback&Deduct)    -1(Rollback&Deduct)(无增的记录，空回滚)
//==============================================================================

// DeductionWallet
// 扣减虚拟币并记录交易
// 1 查询交易
// 1.1 交易存在，并状态一致，为保证幂等性，不做操作，直接返回
// 1.2 如果是正常操作，需要校验是否有同等的回滚操作已执行，如有则直接报错返回！！！！
// 1.3 如果是回滚操作，需要确认之前是否执行过加的操作，未执行过加直接结束
// 2 获取钱包虚拟币信息，校验虚拟币
// 2.1 如果是正常操作，需要校验虚拟币是否充足 (官方账号除外)
// 2.2 如果是回滚操作，支持将虚拟币回滚到负数
// 3 进行扣减余额操作
func DeductionWallet(ctx context.Context, walletId, tradeId, amount int64, coinType basic.CoinType, tradeScene basic.TradeScene, tradeStatus basic.TradeRecordStatus, changeType basic.ChangeType, comment string) error {
	tradeType := getTradeTypeWithStatus(basic.TradeTypeDeduct, tradeStatus)
	//钱包查询和创建放在最外面，提高并发性能
	walletBag, err := getWalletBagDefaultCreate(ctx, walletId, coinType)
	if err != nil {
		return err
	}
	//如果是正常操作，需要校验是否有足够的金额进行扣减，官方钱包账号除外，如果是回滚操作，支持扣减到负数
	if tradeStatus == basic.TradeRecordStatusNormal && !basic.IsOfficialWallet(walletId) {
		if walletBag.Amount < amount {
			return basic.InsufficientAmountErr
		}
	}
	originRecord, err := GetTradeRecord(ctx, walletId, tradeId, coinType, tradeScene, tradeType, changeType)
	if err != nil {
		return err
	}
	if originRecord != nil && originRecord.TradeStatus == tradeStatus {
		//该操作已完成，直接幂等结束
		return nil
	}
	err = WalletDBTX(ctx, func(ctx context.Context, db *gorm.DB) error {
		if originRecord == nil {
			if tradeStatus == basic.TradeRecordStatusRollback {
				//如果是回滚操作，需要确认之前是否执行过加的操作，未执行过加直接结束
				tradeRecord := assembleTradeRecord(tradeId, walletId, amount, tradeScene, basic.TradeRecordStatusEmptyRollback, tradeType, changeType, coinType, comment)
				if err = CreateTradeRecord(ctx, &tradeRecord, db); err != nil {
					return err
				}
				return nil
			}
			//写订单记录
			tradeRecord := assembleTradeRecord(tradeId, walletId, amount, tradeScene, tradeStatus, tradeType, changeType, coinType, comment)
			if err = CreateTradeRecord(ctx, &tradeRecord, db); err != nil {
				return err
			}
		}
		if tradeStatus == basic.TradeRecordStatusNormal {
			//如果是正常操作，需要校验是否有同等的回滚操作已执行，如有则直接报错返回！！！！
			if originRecord != nil && (originRecord.TradeStatus == basic.TradeRecordStatusRollback || originRecord.TradeStatus == basic.TradeRecordStatusEmptyRollback) {
				//之前有同等的回滚操作已执行，直接返回error
				return basic.StateMutationErr
			}
		}
		//更新订单
		if originRecord != nil {
			affect, err := UpdateTradeRecord(ctx, walletId, tradeId, coinType, tradeScene, tradeType, tradeStatus, basic.TradeRecordStatusNormal, changeType, db)
			if err != nil {
				return err
			}
			if !affect {
				//该操作已完成，直接结束
				return nil
			}
		}
		//订单写入/更新成功，对钱包进行操作
		if amount == 0 {
			//如果操作的是0元，直接结束(一般用于某些官方账号加0操作，只记录交易不加钱)
			return nil
		}
		return deductWalletAmount(ctx, walletId, amount, coinType, tradeStatus, db)
	})
	if err != nil {
		return err
	}
	return nil
}

// IncreaseWallet
// 增加虚拟币并记录交易
// 1 查询交易
// 1.1 交易存在，并状态一致，为保证幂等性，不做操作，直接返回
// 1.2 如果是正常操作，需要校验是否有同等的回滚操作已执行，如有则直接报错返回！！！！
// 1.3 如果是回滚操作，需要确认之前是否执行过减的操作，未执行过减直接结束
// 2 获取钱包虚拟币信息，校验虚拟币
// 3 进行增加余额操作
func IncreaseWallet(ctx context.Context, walletId, tradeId, amount int64, coinType basic.CoinType, tradeScene basic.TradeScene, tradeStatus basic.TradeRecordStatus, changeType basic.ChangeType, comment string) error {
	tradeType := getTradeTypeWithStatus(basic.TradeTypeAdd, tradeStatus)
	//不存在则创建放到最外面，提高并发性能
	_, err := getWalletBagDefaultCreate(ctx, walletId, coinType)
	if err != nil {
		return err
	}
	originRecord, err := GetTradeRecord(ctx, walletId, tradeId, coinType, tradeScene, tradeType, changeType)
	if err != nil {
		return err
	}
	if originRecord != nil && originRecord.TradeStatus == tradeStatus {
		//该操作已完成，直接幂等结束
		return nil
	}
	err = WalletDBTX(ctx, func(ctx context.Context, db *gorm.DB) error {
		if originRecord == nil {
			if tradeStatus == basic.TradeRecordStatusRollback {
				//如果是回滚操作，需要确认之前是否执行过减的操作，未执行过减直接结束
				tradeRecord := assembleTradeRecord(tradeId, walletId, amount, tradeScene, basic.TradeRecordStatusEmptyRollback, tradeType, changeType, coinType, comment)
				if err = CreateTradeRecord(ctx, &tradeRecord, db); err != nil {
					return err
				}
				return nil
			}
			tradeRecord := assembleTradeRecord(tradeId, walletId, amount, tradeScene, tradeStatus, tradeType, changeType, coinType, comment)
			if err = CreateTradeRecord(ctx, &tradeRecord, db); err != nil {
				return err
			}
		}
		if tradeStatus == basic.TradeRecordStatusNormal {
			//如果是正常操作，需要校验是否有同等的回滚操作已执行，如有则直接报错返回！！！！
			if originRecord != nil && (originRecord.TradeStatus == basic.TradeRecordStatusRollback || originRecord.TradeStatus == basic.TradeRecordStatusEmptyRollback) {
				//之前有同等的回滚操作已执行，直接返回error
				return basic.StateMutationErr
			}
		}
		//更新订单
		if originRecord != nil {
			affect, err := UpdateTradeRecord(ctx, walletId, tradeId, coinType, tradeScene, tradeType, tradeStatus, basic.TradeRecordStatusNormal, changeType, db)
			if err != nil {
				return err
			}
			if !affect {
				//该操作已完成，返回成功
				return nil
			}
		}
		if amount == 0 {
			//如果金额是0，直接成功返回(一般用于某些官方账号加0操作，只记录交易不加钱)
			return nil
		}
		return increaseWalletAmount(ctx, walletId, amount, coinType, db)
	})
	if err != nil {
		return err
	}
	return nil
}

func assembleTradeRecord(tradeId, walletId, amount int64, tradeScene basic.TradeScene, tradeStatus basic.TradeRecordStatus, tradeType basic.TradeType, changeType basic.ChangeType, coinType basic.CoinType, comment string) model.TradeRecord {
	return model.TradeRecord{
		TradeId:     tradeId,
		TradeScene:  tradeScene,
		WalletId:    walletId,
		Amount:      amount,
		CoinType:    coinType,
		TradeStatus: tradeStatus,
		TradeType:   tradeType,
		ChangeType:  changeType,
		Comment:     comment,
	}
}
