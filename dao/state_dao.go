package dao

import (
	"context"

	"gorm.io/gorm"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"
)

// GetOrCreateState 获取转移记录，如果不存在则创建
func GetOrCreateState(ctx context.Context, req *model.TransferReq) (*model.State, error) {
	var state *model.State
	err := StateInstanceTX(ctx, req.TransferId, func(ctx context.Context, db *gorm.DB) error {
		var err error
		state, err = GetState(ctx, req.TransferId, req.TransferScene, db)
		if err != nil {
			return basic.NewDBFailed(err)
		}
		if state == nil {
			state = model.AssembleState(req.FromAccounts, req.ToAccounts, req.TransferId, req.TransferScene, basic.StateStatusDoing, req.Comment)
			if err = CreateState(ctx, state, db); err != nil {
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

// GetState 获取转移记录
func GetState(ctx context.Context, transferId int64, transferScene basic.TransferScene, db *gorm.DB) (*model.State, error) {
	var records []*model.State
	if db == nil {
		db = basic.GetStateWriteDB(ctx, transferId)
	}
	err := db.Table(model.GetStateTableName(transferId)).
		Where("transfer_id = ? and transfer_scene = ?", transferId, transferScene).
		Find(&records).Error
	if err != nil {
		return nil, basic.NewDBFailed(err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

// UpdateStateStatus 更新转移状态
func UpdateStateStatus(ctx context.Context, transferId int64, transferScene basic.TransferScene, fromStatus, toStatus basic.StateStatus) error {
	err := basic.GetStateWriteDB(ctx, transferId).Table(model.GetStateTableName(transferId)).
		Where("transfer_id = ? and transfer_scene = ? and status = ?", transferId, transferScene, fromStatus).
		Updates(map[string]interface{}{
			"status": toStatus,
		}).Error
	if err != nil {
		return basic.NewDBFailed(err)
	}
	return nil
}

// UpdateStateStatusWithAffect 更新转移状态并返回是否有更改
func UpdateStateStatusWithAffect(ctx context.Context, transferId int64, transferScene basic.TransferScene, fromStatus, toStatus basic.StateStatus) (bool, error) {
	res := basic.GetStateWriteDB(ctx, transferId).Table(model.GetStateTableName(transferId)).
		Where("transfer_id = ? and transfer_scene = ? and status = ?", transferId, transferScene, fromStatus).
		Updates(map[string]interface{}{
			"status": toStatus,
		})
	if res.Error != nil {
		return false, basic.NewDBFailed(res.Error)
	}
	return res.RowsAffected != 0, nil
}

// UpdateStateToRollbackDoing 将非回滚成功的转移状态更新为回滚中
func UpdateStateToRollbackDoing(ctx context.Context, transferId int64, transferScene basic.TransferScene) (bool, error) {
	res := basic.GetStateWriteDB(ctx, transferId).Table(model.GetStateTableName(transferId)).
		Where("transfer_id = ? and transfer_scene = ? and status != ?", transferId, transferScene, basic.StateStatusRollbackDone).
		Updates(map[string]interface{}{
			"status": basic.StateStatusRollbackDoing,
		})
	if res.Error != nil {
		return false, basic.NewDBFailed(res.Error)
	}
	return res.RowsAffected != 0, nil
}

// GetNeedInspectionStateList 获取截止lastTime需要推进的转移记录
func GetNeedInspectionStateList(ctx context.Context, lastTime int64) ([]*model.State, error) {
	var records []*model.State
	for i := int64(0); i < basic.GetDBNum(); i++ {
		for j := int64(0); j < basic.GetStateTableSplitNum(); j++ {
			tableName := model.GetStateTableName(j)
			recordsTmp, err := getLastTimeNeedInspectionStateListByTable(basic.GetStateWriteDB(ctx, i), tableName, lastTime)
			if err != nil {
				return nil, err
			}
			records = append(records, recordsTmp...)
		}
	}
	return records, nil
}

func getLastTimeNeedInspectionStateListByTable(db *gorm.DB, tableName string, lastTime int64) ([]*model.State, error) {
	var records []*model.State
	err := db.Table(tableName).
		Where("status <= ? and updated_at <= ?", basic.StateStatusHalfSuccess, lastTime).
		Find(&records).Error
	if err != nil {
		return nil, basic.NewDBFailed(err)
	}
	return records, nil
}

func CreateState(ctx context.Context, state *model.State, db *gorm.DB) error {
	if db == nil {
		db = basic.GetStateWriteDB(ctx, state.TransferId)
	}
	err := db.Table(model.GetStateTableName(state.TransferId)).Create(state).Error
	if err != nil {
		return basic.NewDBFailed(err)
	}
	return nil
}
