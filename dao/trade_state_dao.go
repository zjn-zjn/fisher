package dao

import (
	"context"

	"github.com/zjn-zjn/coin-trade/basic"
	"github.com/zjn-zjn/coin-trade/model"
	"gorm.io/gorm"
)

// GetOrCreateTradeState 获取交易记录，如果不存在则创建
func GetOrCreateTradeState(ctx context.Context, req *model.CoinTradeReq) (*model.TradeState, error) {
	var state *model.TradeState
	err := WalletDBTX(ctx, func(ctx context.Context, db *gorm.DB) error {
		var err error
		state, err = GetTradeState(ctx, req.TradeId, req.TradeScene, db)
		if err != nil {
			return basic.NewDBFailed(err)
		}
		if state == nil {
			state = model.AssembleTradeState(req.FromWallets, req.ToWallets, req.TradeId, req.TradeScene, basic.TradeStateStatusDoing, req.CoinType, req.Comment)
			if err = CreateTradeState(ctx, state, db); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

// GetTradeState 获取交易记录
func GetTradeState(ctx context.Context, tradeId int64, tradeScene basic.TradeScene, db *gorm.DB) (*model.TradeState, error) {
	var records []*model.TradeState
	if db == nil {
		db = basic.GetCoinTradeWriteDB(ctx)
	}
	err := db.Table(model.GetTradeStateTableName(tradeId)).
		Where("trade_id = ? and trade_scene = ?", tradeId, tradeScene).
		Find(&records).Error
	if err != nil {
		return nil, basic.NewDBFailed(err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

// UpdateTradeStateStatus 更新交易状态
func UpdateTradeStateStatus(ctx context.Context, tradeId int64, tradeScene basic.TradeScene, fromStatus, toStatus basic.TradeStateStatus) error {
	err := basic.GetCoinTradeWriteDB(ctx).Table(model.GetTradeStateTableName(tradeId)).
		Where("trade_id = ? and trade_scene = ? and status = ?", tradeId, tradeScene, fromStatus).
		Updates(map[string]interface{}{
			"status": toStatus,
		}).Error
	if err != nil {
		return basic.NewDBFailed(err)
	}
	return nil
}

// UpdateTradeStateStatusWithAffect 更新交易状态并返回是否有更改
func UpdateTradeStateStatusWithAffect(ctx context.Context, tradeId int64, tradeScene basic.TradeScene, fromStatus, toStatus basic.TradeStateStatus) (bool, error) {
	res := basic.GetCoinTradeWriteDB(ctx).Table(model.GetTradeStateTableName(tradeId)).
		Where("trade_id = ? and trade_scene = ? and status = ?", tradeId, tradeScene, fromStatus).
		Updates(map[string]interface{}{
			"status": toStatus,
		})
	if res.Error != nil {
		return false, basic.NewDBFailed(res.Error)
	}
	return res.RowsAffected != 0, nil
}

// UpdateTradeStateToRollbackDoing 将非回滚成功的交易状态更新为回滚中
func UpdateTradeStateToRollbackDoing(ctx context.Context, tradeId int64, tradeScene basic.TradeScene) (bool, error) {
	res := basic.GetCoinTradeWriteDB(ctx).Table(model.GetTradeStateTableName(tradeId)).
		Where("trade_id = ? and trade_scene = ? and status != ?", tradeId, tradeScene, basic.TradeStateStatusRollbackDone).
		Updates(map[string]interface{}{
			"status": basic.TradeStateStatusRollbackDoing,
		})
	if res.Error != nil {
		return false, basic.NewDBFailed(res.Error)
	}
	return res.RowsAffected != 0, nil
}

// GetNeedInspectionTradeStateList 获取截止lastTime需要推进的交易记录
func GetNeedInspectionTradeStateList(ctx context.Context, lastTime int64) ([]*model.TradeState, error) {
	if basic.GetTradeStateTableSplitNum() <= 1 {
		return getLastTimeNeedInspectionTradeStateListByTable(ctx, model.TradeStateTablePrefix, lastTime)
	}
	var records []*model.TradeState
	for i := int64(0); i < basic.GetTradeStateTableSplitNum(); i++ {
		tableName := model.GetTradeStateTableName(i)
		recordsTmp, err := getLastTimeNeedInspectionTradeStateListByTable(ctx, tableName, lastTime)
		if err != nil {
			return nil, err
		}
		records = append(records, recordsTmp...)
	}
	return records, nil
}

func getLastTimeNeedInspectionTradeStateListByTable(ctx context.Context, tableName string, lastTime int64) ([]*model.TradeState, error) {
	var records []*model.TradeState
	err := basic.GetCoinTradeWriteDB(ctx).Table(tableName).
		Where("status <= ? and updated_at <= ?", basic.TradeStateStatusHalfSuccess, lastTime).
		Find(&records).Error
	if err != nil {
		return nil, basic.NewDBFailed(err)
	}
	return records, nil
}

func CreateTradeState(ctx context.Context, state *model.TradeState, db *gorm.DB) error {
	if db == nil {
		db = basic.GetCoinTradeWriteDB(ctx)
	}
	err := db.Table(model.GetTradeStateTableName(state.TradeId)).Create(state).Error
	if err != nil {
		return basic.NewDBFailed(err)
	}
	return nil
}
