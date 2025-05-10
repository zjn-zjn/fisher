package dao

import (
	"context"
	"errors"

	"github.com/zjn-zjn/fisher/model"

	"gorm.io/gorm"

	"github.com/zjn-zjn/fisher/basic"
)

// GetRecord 获取转移记录
func GetRecord(ctx context.Context, accountId, transferId int64, itemType basic.ItemType, transferScene basic.TransferScene, transferType basic.TransferType, changeType basic.ChangeType) (*model.Record, error) {
	var records []model.Record
	if err := basic.GetRecordAndAccountWriteDB(ctx, accountId).Table(model.GetRecordTableName(accountId)).
		Where("account_id = ? and transfer_id = ? and  item_type = ? and transfer_scene = ? and transfer_type = ? and change_type = ?", accountId, transferId, itemType, transferScene, transferType, changeType).
		Find(&records).Error; err != nil {
		return nil, basic.NewDBFailed(err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

// UpdateRecord 更新转移记录
func UpdateRecord(ctx context.Context, accountId, transferId int64, itemType basic.ItemType, transferScene basic.TransferScene, transferType basic.TransferType, transferStatus, originTransferStatus basic.RecordStatus, changeType basic.ChangeType, db *gorm.DB) (bool, error) {
	if db == nil {
		db = basic.GetRecordAndAccountWriteDB(ctx, accountId)
	}
	result := db.Table(model.GetRecordTableName(accountId)).
		Where("account_id = ? and transfer_id = ? and  item_type = ? and transfer_scene = ? and transfer_type = ? and transfer_status = ? and change_type = ?", accountId, transferId, itemType, transferScene, transferType, originTransferStatus, changeType).
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
		db = basic.GetRecordAndAccountWriteDB(ctx, record.AccountId)
	}
	if err := db.Table(model.GetRecordTableName(record.AccountId)).Create(record).Error; err != nil {
		return basic.NewDBFailed(err)
	}
	return nil
}

func GetAccountLastRecord(ctx context.Context, accountId int64, itemType *basic.ItemType, transferScene *basic.TransferScene, transferType *basic.TransferType, db *gorm.DB) (*model.Record, error) {
	if db == nil {
		db = basic.GetRecordAndAccountReadDB(ctx, accountId)
	}
	var record model.Record
	db = db.Table(model.GetRecordTableName(accountId)).Where("account_id = ?", accountId)
	if itemType != nil {
		db = db.Where("item_type = ?", *itemType)
	}
	if transferScene != nil {
		db = db.Where("transfer_scene = ?", *transferScene)
	}
	if transferType != nil {
		db = db.Where("transfer_type = ?", *transferType)
	}
	if err := db.Where(`transfer_status = ?`, basic.RecordStatusNormal).Order("id desc").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}
