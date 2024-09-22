package model

import "github.com/zjn-zjn/fisher/basic"

const (
	BagTablePrefix = "bag"
)

type Bag struct {
	ID        int64          `json:"id" gorm:"column:id;"`                                     // 主键
	BagId     int64          `json:"bag_id" gorm:"column:bag_id;"`                             // 背包ID
	ItemType  basic.ItemType `json:"item_type" gorm:"column:item_type;"`                       // 转移物品类型
	Amount    int64          `json:"amount" gorm:"column:amount;"`                             // 数量
	CreatedAt int64          `json:"created_at" gorm:"column:created_at;autoCreateTime:milli"` // 创建时间
	UpdatedAt int64          `json:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli"` // 创建时间
}

func GetBagTableName(bagId int64) string {
	return BagTablePrefix + basic.GetBagTableSuffix(bagId)
}
