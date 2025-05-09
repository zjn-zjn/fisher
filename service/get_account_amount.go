package service

import (
	"context"
	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/dao"
)

func GetAccountAmountRead(ctx context.Context, accountId int64) (map[basic.ItemType]int64, error) {
	return dao.GetAccountAmount(ctx, accountId, basic.GetAccountReadDB(ctx, accountId))
}

func GetAccountAmountWrite(ctx context.Context, accountId int64) (map[basic.ItemType]int64, error) {
	return dao.GetAccountAmount(ctx, accountId, basic.GetAccountWriteDB(ctx, accountId))
}

func GetAccountAmountByItemTypeRead(ctx context.Context, accountId int64, itemType basic.ItemType) (int64, error) {
	return dao.GetAccountAmountByItemType(ctx, accountId, itemType, basic.GetAccountReadDB(ctx, accountId))
}

func GetAccountAmountByItemTypeWrite(ctx context.Context, accountId int64, itemType basic.ItemType) (int64, error) {
	return dao.GetAccountAmountByItemType(ctx, accountId, itemType, basic.GetRecordAndAccountWriteDB(ctx, accountId))
}
