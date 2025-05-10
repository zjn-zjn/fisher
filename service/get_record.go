package service

import (
	"context"
	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/dao"
	"github.com/zjn-zjn/fisher/model"
)

func GetAccountLastRecordRead(ctx context.Context, accountId int64, itemType *basic.ItemType, transferScene *basic.TransferScene, transferType *basic.TransferType) (*model.Record, error) {
	return dao.GetAccountLastRecord(ctx, accountId, itemType, transferScene, transferType, basic.GetRecordAndAccountReadDB(ctx, accountId))
}

func GetAccountLastRecordWrite(ctx context.Context, accountId int64, itemType *basic.ItemType, transferScene *basic.TransferScene, transferType *basic.TransferType) (*model.Record, error) {
	return dao.GetAccountLastRecord(ctx, accountId, itemType, transferScene, transferType, basic.GetRecordAndAccountWriteDB(ctx, accountId))
}
