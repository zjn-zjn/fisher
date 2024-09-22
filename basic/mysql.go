package basic

import (
	"context"

	"gorm.io/plugin/dbresolver"

	"gorm.io/gorm"
)

var fisherDB *gorm.DB

// initItemTransferDB 初始化物品转移数据库
func initItemTransferDB(db *gorm.DB) {
	fisherDB = db
}

func GetWriteDB(ctx context.Context) *gorm.DB {
	return fisherDB.Clauses(dbresolver.Write).WithContext(ctx)
}

func GetReadDB(ctx context.Context) *gorm.DB {
	return fisherDB.Clauses(dbresolver.Read).WithContext(ctx)
}
