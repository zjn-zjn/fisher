package dao

import (
	"context"

	"github.com/zjn-zjn/fisher/model"

	"gorm.io/gorm"

	"github.com/zjn-zjn/fisher/basic"
)

// GetRecord 获取转移记录
func GetRecord(ctx context.Context, bagId, transferId int64, itemType basic.ItemType, transferScene basic.TransferScene, transferType basic.RecordType, changeType basic.ChangeType) (*model.Record, error) {
	var records []model.Record
	if err := basic.GetWriteDB(ctx).Table(model.GetRecordTableName(bagId)).
		Where("bag_id = ? and transfer_id = ? and  item_type = ? and transfer_scene = ? and transfer_type = ? and change_type = ?", bagId, transferId, itemType, transferScene, transferType, changeType).
		Find(&records).Error; err != nil {
		return nil, basic.NewDBFailed(err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

// UpdateRecord 更新转移记录
func UpdateRecord(ctx context.Context, bagId, transferId int64, itemType basic.ItemType, transferScene basic.TransferScene, transferType basic.RecordType, transferStatus, originTransferStatus basic.RecordStatus, changeType basic.ChangeType, db *gorm.DB) (bool, error) {
	if db == nil {
		db = basic.GetWriteDB(ctx)
	}
	result := db.Table(model.GetRecordTableName(bagId)).
		Where("bag_id = ? and transfer_id = ? and  item_type = ? and transfer_scene = ? and transfer_type = ? and transfer_status = ? and change_type = ?", bagId, transferId, itemType, transferScene, transferType, originTransferStatus, changeType).
		Update("transfer_status", transferStatus)
	if err := result.Error; err != nil {
		return false, basic.NewDBFailed(err)
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func CreateRecord(ctx context.Context, record *model.Record, db *gorm.DB) error {
	if db == nil {
		db = basic.GetWriteDB(ctx)
	}
	if err := db.Table(model.GetRecordTableName(record.BagId)).Create(record).Error; err != nil {
		return basic.NewDBFailed(err)
	}
	return nil
}
