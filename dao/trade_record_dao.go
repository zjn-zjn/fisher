package dao

import (
	"context"

	"github.com/zjn-zjn/coin-trade/model"

	"github.com/zjn-zjn/coin-trade/conf"

	"gorm.io/gorm"

	"github.com/zjn-zjn/coin-trade/basic"
)

// GetTradeRecord 获取交易记录
func GetTradeRecord(ctx context.Context, walletId, tradeId int64, coinType basic.CoinType, tradeScene basic.TradeScene, tradeType basic.TradeType, changeType basic.ChangeType, db *gorm.DB) (*model.TradeRecord, error) {
	var records []model.TradeRecord
	if db == nil {
		db = conf.GetCoinTradeWriteDB(ctx)
	}
	if err := db.Table(model.GetTradeRecordTableName(walletId)).
		Where("wallet_id = ? and trade_id = ? and  coin_type = ? and trade_scene = ? and trade_type = ? and change_type = ?", walletId, tradeId, coinType, tradeScene, tradeType, changeType).
		Find(&records).Error; err != nil {
		return nil, basic.NewDBFailed(err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

// UpdateTradeRecord 更新交易记录
func UpdateTradeRecord(ctx context.Context, walletId, tradeId int64, coinType basic.CoinType, tradeScene basic.TradeScene, tradeType basic.TradeType, tradeStatus, originTradeStatus basic.TradeRecordStatus, changeType basic.ChangeType, db *gorm.DB) (bool, error) {
	if db == nil {
		db = conf.GetCoinTradeWriteDB(ctx)
	}
	result := db.Table(model.GetTradeRecordTableName(walletId)).
		Where("wallet_id = ? and trade_id = ? and  coin_type = ? and trade_scene = ? and trade_type = ? and trade_status = ? and change_type = ?", walletId, tradeId, coinType, tradeScene, tradeType, originTradeStatus, changeType).
		Update("trade_status", tradeStatus)
	if err := result.Error; err != nil {
		return false, basic.NewDBFailed(err)
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func CreateTradeRecord(ctx context.Context, tradeRecord *model.TradeRecord, db *gorm.DB) error {
	if db == nil {
		db = conf.GetCoinTradeWriteDB(ctx)
	}
	if err := db.Table(model.GetTradeRecordTableName(tradeRecord.WalletId)).Create(tradeRecord).Error; err != nil {
		return basic.NewDBFailed(err)
	}
	return nil
}
