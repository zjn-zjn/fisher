package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/zjn-zjn/fisher/basic"
)

const (
	StateTablePrefix = "state"
)

type State struct {
	ID            int64               `json:"id" gorm:"column:id;"`                                     // 主键
	TransferId    int64               `json:"transfer_id" gorm:"column:transfer_id;"`                   // 转移ID
	TransferScene basic.TransferScene `json:"transfer_scene" gorm:"column:transfer_scene;"`             // 转移场景
	FromBags      BagList             `json:"from_bags" gorm:"column:from_bags;"`                       // 扣款背包信息列表
	ToBags        BagList             `json:"to_bags" gorm:"column:to_bags;"`                           // 收款背包信息列表
	ItemType      basic.ItemType      `json:"item_type" gorm:"column:item_type;"`                       // 物品类型
	Status        basic.StateStatus   `json:"status" gorm:"column:status;"`                             // 转移状态 1-进行中 2-回滚中 3-半成功 4-成功 5-已回滚
	Comment       string              `json:"comment" gorm:"column:comment;"`                           // 转移备注
	CreatedAt     int64               `json:"created_at" gorm:"column:created_at;autoCreateTime:milli"` // 创建时间
	UpdatedAt     int64               `json:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli"` // 创建时间
}

type BagList []*TransferItem

func GetStateTableName(transferId int64) string {
	if basic.GetStateTableSplitNum() == 1 {
		return StateTablePrefix
	}
	return StateTablePrefix + basic.GetStateTableSuffix(transferId)
}

func AssembleState(fromBags, toBags []*TransferItem, transferId int64, transferScene basic.TransferScene, status basic.StateStatus, itemType basic.ItemType, comment string) *State {
	md := &State{
		TransferId:    transferId,
		TransferScene: transferScene,
		ItemType:      itemType,
		Status:        status,
		Comment:       comment,
	}
	if fromBags == nil {
		md.FromBags = make(BagList, 0)
	} else {
		md.FromBags = fromBags
	}
	if toBags == nil {
		md.ToBags = make(BagList, 0)
	} else {
		md.ToBags = toBags
	}
	return md
}

func (m *BagList) Scan(val interface{}) error {
	s := val.([]uint8)
	var toBags BagList
	err := json.Unmarshal(s, &toBags)
	if err != nil {
		return err
	}
	*m = toBags
	return nil
}

func (m BagList) Value() (driver.Value, error) {
	result, _ := json.Marshal(m)
	return string(result), nil
}
