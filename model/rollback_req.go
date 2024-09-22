package model

import "github.com/zjn-zjn/fisher/basic"

type RollbackReq struct {
	TransferId    int64               `json:"transfer_id"`
	TransferScene basic.TransferScene `json:"transfer_scene"`
}
