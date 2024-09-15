package basic

import (
	"context"

	"gorm.io/plugin/dbresolver"

	"gorm.io/gorm"
)

var coinTradeDB *gorm.DB

// initCoinTradeDB 初始化虚拟币交易数据库
func initCoinTradeDB(db *gorm.DB) {
	coinTradeDB = db
}

func GetCoinTradeWriteDB(ctx context.Context) *gorm.DB {
	return coinTradeDB.Clauses(dbresolver.Write).WithContext(ctx)
}

func GetCoinTradeReadDB(ctx context.Context) *gorm.DB {
	return coinTradeDB.Clauses(dbresolver.Read).WithContext(ctx)
}
