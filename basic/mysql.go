package basic

import (
	"context"

	"gorm.io/plugin/dbresolver"

	"gorm.io/gorm"
)

var fisherDBs []*gorm.DB

// initItemTransferDB 初始化物品转移数据库
func initItemTransferDB(dbs []*gorm.DB) {
	fisherDBs = dbs
	dbNum = int64(len(dbs))
}

func GetStateWriteDB(ctx context.Context, transferId int64) *gorm.DB {
	idx := 0
	if dbNum >= 1 {
		idx = int(transferId % dbNum)
	}
	return fisherDBs[idx].Clauses(dbresolver.Write).WithContext(ctx)
}

func GetRecordAndAccountWriteDB(ctx context.Context, accountId int64) *gorm.DB {
	idx := 0
	if dbNum >= 1 {
		idx = int(accountId % dbNum)
	}
	return fisherDBs[idx].Clauses(dbresolver.Write).WithContext(ctx)
}

func GetAccountWriteDB(ctx context.Context, accountId int64) *gorm.DB {
	idx := 0
	if dbNum >= 1 {
		idx = int(accountId % dbNum)
	}
	return fisherDBs[idx].Clauses(dbresolver.Write).WithContext(ctx)
}

func GetAccountReadDB(ctx context.Context, accountId int64) *gorm.DB {
	idx := 0
	if dbNum >= 1 {
		idx = int(accountId % dbNum)
	}
	return fisherDBs[idx].Clauses(dbresolver.Read).WithContext(ctx)
}

func GetStateReadDB(ctx context.Context, transferId int64) *gorm.DB {
	idx := 0
	if dbNum >= 1 {
		idx = int(transferId % dbNum)
	}
	return fisherDBs[idx].Clauses(dbresolver.Read).WithContext(ctx)
}

func GetRecordAndAccountReadDB(ctx context.Context, accountId int64) *gorm.DB {
	idx := 0
	if dbNum >= 1 {
		idx = int(accountId % dbNum)
	}
	return fisherDBs[idx].Clauses(dbresolver.Read).WithContext(ctx)
}
