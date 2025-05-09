package model

import "github.com/zjn-zjn/fisher/basic"

type TransferReq struct {
	TransferId     int64               `json:"transfer_id"`      // 转移ID
	UseHalfSuccess bool                `json:"use_half_success"` // 是否使用半成功，适用于扣减成功即可认为转移成功的转移场景，增加操作即使失败也会尝试持续推进至成功
	FromAccounts   []*TransferItem     `json:"from_accounts"`    // 转移发起者
	ToAccounts     []*TransferItem     `json:"to_accounts"`      // 转移接收者
	TransferScene  basic.TransferScene `json:"transfer_scene"`   // 转移场景
	Comment        string              `json:"comment"`          // 转移备注
}

type TransferItem struct {
	AccountId  int64            `json:"account_id"`  // 转移接收者ID
	ItemType   basic.ItemType   `json:"item_type"`   // 转移币种
	Amount     int64            `json:"amount"`      // 接收金额
	ChangeType basic.ChangeType `json:"change_type"` // 变更类型
	Comment    string           `json:"comment"`     // 转移备注
}
