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
	FromAccounts  AccountList         `json:"from_accounts" gorm:"column:from_accounts;"`               // 扣款账户信息列表
	ToAccounts    AccountList         `json:"to_accounts" gorm:"column:to_accounts;"`                   // 收款账户信息列表
	Status        basic.StateStatus   `json:"status" gorm:"column:status;"`                             // 转移状态 1-进行中 2-回滚中 3-半成功 4-成功 5-已回滚
	Comment       string              `json:"comment" gorm:"column:comment;"`                           // 转移备注
	CreatedAt     int64               `json:"created_at" gorm:"column:created_at;autoCreateTime:milli"` // 创建时间
	UpdatedAt     int64               `json:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli"` // 创建时间
}

type AccountList []*TransferItem

func GetStateTableName(transferId int64) string {
	if basic.GetStateTableSplitNum() == 1 {
		return StateTablePrefix
	}
	return StateTablePrefix + basic.GetStateTableSuffix(transferId)
}

func AssembleState(fromAccounts, toAccounts []*TransferItem, transferId int64, transferScene basic.TransferScene, status basic.StateStatus, comment string) *State {
	md := &State{
		TransferId:    transferId,
		TransferScene: transferScene,
		Status:        status,
		Comment:       comment,
	}
	if fromAccounts == nil {
		md.FromAccounts = make(AccountList, 0)
	} else {
		md.FromAccounts = fromAccounts
	}
	if toAccounts == nil {
		md.ToAccounts = make(AccountList, 0)
	} else {
		md.ToAccounts = toAccounts
	}
	return md
}

func (m *AccountList) Scan(val interface{}) error {
	s := val.([]uint8)
	var toAccounts AccountList
	err := json.Unmarshal(s, &toAccounts)
	if err != nil {
		return err
	}
	*m = toAccounts
	return nil
}

func (m AccountList) Value() (driver.Value, error) {
	result, _ := json.Marshal(m)
	return string(result), nil
}
