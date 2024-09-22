package model

import "github.com/zjn-zjn/fisher/basic"

const (
	RecordTablePrefix = "record"
)

type Record struct {
	ID             int64               `json:"id" gorm:"column:id;"`                                     // 主键
	BagId          int64               `json:"bag_id" gorm:"column:bag_id;"`                             // 背包ID
	TransferId     int64               `json:"transfer_id" gorm:"column:transfer_id;"`                   // 转移ID
	TransferScene  basic.TransferScene `json:"transfer_scene" gorm:"column:transfer_scene;"`             // 转移场景
	RecordType     basic.RecordType    `json:"transfer_type" gorm:"column:transfer_type;"`               // 转移类型
	TransferStatus basic.RecordStatus  `json:"transfer_status" gorm:"column:transfer_status;"`           // 转移状态
	Amount         int64               `json:"amount" gorm:"column:amount;"`                             // 转移金额
	ItemType       basic.ItemType      `json:"item_type" gorm:"column:item_type;"`                       // 转移币种
	ChangeType     basic.ChangeType    `json:"change_type" gorm:"column:change_type;"`                   // 转移变化类型
	Comment        string              `json:"comment" gorm:"column:comment;"`                           // 转移备注
	CreatedAt      int64               `json:"created_at" gorm:"column:created_at;autoCreateTime:milli"` // 创建时间
	UpdatedAt      int64               `json:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli"` // 创建时间
}

func GetRecordTableName(bagId int64) string {
	return RecordTablePrefix + basic.GetRecordTableSuffix(bagId)
}
