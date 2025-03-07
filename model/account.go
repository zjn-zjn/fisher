package model

import "github.com/zjn-zjn/fisher/basic"

const (
	AccountTablePrefix = "account"
)

type Account struct {
	ID        int64          `json:"id" gorm:"column:id;"`                                     // 主键
	AccountId int64          `json:"account_id" gorm:"column:account_id;"`                     // 账户ID
	ItemType  basic.ItemType `json:"item_type" gorm:"column:item_type;"`                       // 转移物品类型
	Amount    int64          `json:"amount" gorm:"column:amount;"`                             // 数量
	CreatedAt int64          `json:"created_at" gorm:"column:created_at;autoCreateTime:milli"` // 创建时间
	UpdatedAt int64          `json:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli"` // 创建时间
}

func GetAccountTableName(accountId int64) string {
	return AccountTablePrefix + basic.GetAccountTableSuffix(accountId)
}
