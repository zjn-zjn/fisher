package dao

import (
	"context"

	"github.com/go-sql-driver/mysql"

	"github.com/zjn-zjn/coin-trade/basic"
	"github.com/zjn-zjn/coin-trade/conf"
	"github.com/zjn-zjn/coin-trade/model"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func deductWalletAmount(ctx context.Context, walletId, amount int64, coinType basic.CoinType, tradeStatus basic.TradeRecordStatus, db *gorm.DB) error {
	if db == nil {
		db = conf.GetCoinTradeWriteDB(ctx)
	}
	var walletDB = db.Table(model.GetWalletBagTableName(walletId))
	//这里采用update coin = coin - 1 where coin - amount >= 0 的方式进行扣减，提高并发性能
	if tradeStatus == basic.TradeRecordStatusRollback || basic.IsOfficialWallet(walletId) {
		//官方账号和回滚 不使用coin - amount >= 0条件，直接扣减
		walletDB = walletDB.Where("wallet_id = ? and coin_type = ?", walletId, coinType)
	} else {
		walletDB = walletDB.Where("wallet_id = ?  and coin_type = ? and amount - ? >= 0", walletId, coinType, amount)
	}
	res := walletDB.UpdateColumn("amount", gorm.Expr("amount - ?", amount))
	if res.Error != nil {
		return basic.NewDBFailed(res.Error)
	}
	if res.RowsAffected == 0 {
		//这里理论上只能是由于金额不足引起的，直接返回错误
		return basic.InsufficientAmountErr
	}
	return nil
}

func addWalletAmount(ctx context.Context, walletId, amount int64, coinType basic.CoinType, tradeStatus basic.TradeRecordStatus, db *gorm.DB) error {
	if db == nil {
		db = conf.GetCoinTradeWriteDB(ctx)
	}
	//这里采用update coin = coin + amount 的方式进行增加，提高并发性能
	res := db.Table(model.GetWalletBagTableName(walletId)).
		Where("wallet_id = ? and coin_type = ?", walletId, coinType).
		UpdateColumn("amount", gorm.Expr("amount + ?", amount))
	if res.Error != nil {
		return basic.NewDBFailed(res.Error)
	}
	if res.RowsAffected == 0 {
		//这里理论上不会发生，增加的金额>0，并且成功，理论上不会有0行影响，以防万一，还是加上
		return basic.StateMutationErr
	}
	return nil
}

func getWalletBagDefaultCreate(ctx context.Context, walletId int64, coinType basic.CoinType) (*model.WalletBag, error) {
	var walletBag *model.WalletBag
	var walletBags []model.WalletBag
	//获取钱包，不存在就创建
	if err := conf.GetCoinTradeWriteDB(ctx).Table(model.GetWalletBagTableName(walletId)).
		Where("wallet_id = ? and coin_type = ?", walletId, coinType).
		Find(&walletBags).Error; err != nil {
		return nil, basic.NewDBFailed(err)
	}
	if len(walletBags) != 0 {
		return &walletBags[0], nil
	}
	walletBag = &model.WalletBag{
		WalletId: walletId,
		Amount:   0,
		CoinType: coinType,
	}
	if err := conf.GetCoinTradeWriteDB(ctx).Table(model.GetWalletBagTableName(walletId)).Create(&walletBag).Error; err != nil {
		//如果是唯一键冲突错误，则再次查询
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			if err = conf.GetCoinTradeWriteDB(ctx).Table(model.GetWalletBagTableName(walletId)).
				Where("wallet_id = ? and coin_type = ?", walletId, coinType).
				Find(&walletBags).Error; err != nil {
				return nil, basic.NewDBFailed(err)
			}
			return &walletBags[0], nil
		}
		return nil, basic.NewDBFailed(err)
	}
	return walletBag, nil
}

func getTradeTypeWithStatus(tradeType basic.TradeType, tradeStatus basic.TradeRecordStatus) basic.TradeType {
	if tradeStatus == basic.TradeRecordStatusNormal {
		//正向操作，直接返回
		return tradeType
	}
	//逆向操作，返回相反的操作
	if tradeType == basic.TradeTypeAdd {
		return basic.TradeTypeDeduct
	}
	return basic.TradeTypeAdd
}
