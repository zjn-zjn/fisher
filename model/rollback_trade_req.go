package model

import "github.com/zjn-zjn/coin-trade/basic"

type RollbackTradeReq struct {
	TradeId    int64            `json:"trade_id"`
	TradeScene basic.TradeScene `json:"trade_scene"`
}
