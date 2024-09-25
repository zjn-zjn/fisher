package dao

import (
	"context"

	"github.com/go-sql-driver/mysql"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func deductBagAmount(ctx context.Context, bagId, amount int64, itemType basic.ItemType, transferStatus basic.RecordStatus, db *gorm.DB) error {
	if db == nil {
		db = basic.GetRecordAndBagWriteDB(ctx, bagId)
	}
	var bagDB = db.Table(model.GetBagTableName(bagId))
	//这里采用update item = item - 1 where item - amount >= 0 的方式进行扣减，提高并发成功率
	if transferStatus == basic.RecordStatusRollback || basic.IsOfficialBag(bagId) {
		//官方账号和回滚 不使用item - amount >= 0条件，直接扣减
		bagDB = bagDB.Where("bag_id = ? and item_type = ?", bagId, itemType)
	} else {
		bagDB = bagDB.Where("bag_id = ?  and item_type = ? and amount - ? >= 0", bagId, itemType, amount)
	}
	res := bagDB.UpdateColumn("amount", gorm.Expr("amount - ?", amount))
	if res.Error != nil {
		return basic.NewDBFailed(res.Error)
	}
	if res.RowsAffected == 0 {
		//这里理论上只能是由于金额不足引起的，直接返回错误
		return basic.InsufficientAmountErr
	}
	return nil
}

func increaseBagAmount(ctx context.Context, bagId, amount int64, itemType basic.ItemType, db *gorm.DB) error {
	if db == nil {
		db = basic.GetRecordAndBagWriteDB(ctx, bagId)
	}
	//这里采用update item = item + amount 的方式进行增加，提高并发成功率
	res := db.Table(model.GetBagTableName(bagId)).
		Where("bag_id = ? and item_type = ?", bagId, itemType).
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

func getBagDefaultCreate(ctx context.Context, bagId int64, itemType basic.ItemType) (*model.Bag, error) {
	var bag *model.Bag
	var bags []model.Bag
	//获取背包，不存在就创建
	if err := basic.GetRecordAndBagWriteDB(ctx, bagId).Table(model.GetBagTableName(bagId)).
		Where("bag_id = ? and item_type = ?", bagId, itemType).
		Find(&bags).Error; err != nil {
		return nil, basic.NewDBFailed(err)
	}
	if len(bags) != 0 {
		return &bags[0], nil
	}
	bag = &model.Bag{
		BagId:    bagId,
		Amount:   0,
		ItemType: itemType,
	}
	if err := basic.GetRecordAndBagWriteDB(ctx, bagId).Table(model.GetBagTableName(bagId)).Create(&bag).Error; err != nil {
		//如果是唯一键冲突错误，则再次查询
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			if err = basic.GetRecordAndBagWriteDB(ctx, bagId).Table(model.GetBagTableName(bagId)).
				Where("bag_id = ? and item_type = ?", bagId, itemType).
				Find(&bags).Error; err != nil {
				return nil, basic.NewDBFailed(err)
			}
			return &bags[0], nil
		}
		return nil, basic.NewDBFailed(err)
	}
	return bag, nil
}

func getRecordTypeWithStatus(transferType basic.RecordType, transferStatus basic.RecordStatus) basic.RecordType {
	if transferStatus == basic.RecordStatusNormal {
		//正向操作，直接返回
		return transferType
	}
	//逆向操作，返回相反的操作
	if transferType == basic.RecordTypeAdd {
		return basic.RecordTypeDeduct
	}
	return basic.RecordTypeAdd
}
