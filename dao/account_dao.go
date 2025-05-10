package dao

import (
	"context"

	"github.com/go-sql-driver/mysql"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func GetAccountAmountByItemType(ctx context.Context, accountId int64, itemType basic.ItemType, db *gorm.DB) (int64, error) {
	if db == nil {
		db = basic.GetAccountReadDB(ctx, accountId)
	}
	var account model.Account
	if err := db.Table(model.GetAccountTableName(accountId)).Where("account_id = ? and item_type = ?", accountId, itemType).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return account.Amount, nil
}

func GetAccountAmount(ctx context.Context, accountId int64, db *gorm.DB) (map[basic.ItemType]int64, error) {
	if db == nil {
		db = basic.GetAccountReadDB(ctx, accountId)
	}
	var accounts []model.Account
	if err := db.Table(model.GetAccountTableName(accountId)).Where("account_id = ?", accountId).Find(&accounts).Error; err != nil {
		return nil, err
	}
	amountMap := make(map[basic.ItemType]int64)
	for _, account := range accounts {
		amountMap[account.ItemType] = account.Amount
	}
	return amountMap, nil
}

func deductAccountAmount(ctx context.Context, accountId, amount int64, itemType basic.ItemType, transferStatus basic.RecordStatus, db *gorm.DB) error {
	if db == nil {
		db = basic.GetRecordAndAccountWriteDB(ctx, accountId)
	}
	var accountDB = db.Table(model.GetAccountTableName(accountId))
	//这里采用update item = item - 1 where item - amount >= 0 的方式进行扣减，提高并发成功率
	if transferStatus == basic.RecordStatusRollback || basic.IsOfficialAccount(accountId) {
		//官方账号和回滚 不使用item - amount >= 0条件，直接扣减
		accountDB = accountDB.Where("account_id = ? and item_type = ?", accountId, itemType)
	} else {
		accountDB = accountDB.Where("account_id = ?  and item_type = ? and amount - ? >= 0", accountId, itemType, amount)
	}
	res := accountDB.UpdateColumn("amount", gorm.Expr("amount - ?", amount))
	if res.Error != nil {
		return basic.NewDBFailed(res.Error)
	}
	if res.RowsAffected == 0 {
		//这里理论上只能是由于金额不足引起的，直接返回错误
		return basic.InsufficientAmountErr
	}
	return nil
}

func increaseAccountAmount(ctx context.Context, accountId, amount int64, itemType basic.ItemType, db *gorm.DB) error {
	if db == nil {
		db = basic.GetRecordAndAccountWriteDB(ctx, accountId)
	}
	//这里采用update item = item + amount 的方式进行增加，提高并发成功率
	res := db.Table(model.GetAccountTableName(accountId)).
		Where("account_id = ? and item_type = ?", accountId, itemType).
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

func getAccountDefaultCreate(ctx context.Context, accountId int64, itemType basic.ItemType) (*model.Account, error) {
	var account *model.Account
	var accounts []model.Account
	//获取账户，不存在就创建
	if err := basic.GetRecordAndAccountWriteDB(ctx, accountId).Table(model.GetAccountTableName(accountId)).
		Where("account_id = ? and item_type = ?", accountId, itemType).
		Find(&accounts).Error; err != nil {
		return nil, basic.NewDBFailed(err)
	}
	if len(accounts) != 0 {
		return &accounts[0], nil
	}
	account = &model.Account{
		AccountId: accountId,
		Amount:    0,
		ItemType:  itemType,
	}
	if err := basic.GetRecordAndAccountWriteDB(ctx, accountId).Table(model.GetAccountTableName(accountId)).Create(&account).Error; err != nil {
		//如果是唯一键冲突错误，则再次查询
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			if err = basic.GetRecordAndAccountWriteDB(ctx, accountId).Table(model.GetAccountTableName(accountId)).
				Where("account_id = ? and item_type = ?", accountId, itemType).
				Find(&accounts).Error; err != nil {
				return nil, basic.NewDBFailed(err)
			}
			return &accounts[0], nil
		}
		return nil, basic.NewDBFailed(err)
	}
	return account, nil
}

func getRecordTypeWithStatus(transferType basic.TransferType, transferStatus basic.RecordStatus) basic.TransferType {
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
