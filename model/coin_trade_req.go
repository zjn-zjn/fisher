package model

import "github.com/zjn-zjn/coin-trade/basic"

type CoinTradeReq struct {
	TradeId        int64              `json:"trade_id"`         // 交易ID
	UseHalfSuccess bool               `json:"use_half_success"` // 是否使用半成功，适用于扣减成功即可认为交易成功的交易场景，增加操作即使失败也会尝试持续推进至成功
	CoinType       basic.CoinType     `json:"coin_type"`        // 交易币种
	FromWallets    []*TradeWalletItem `json:"from_wallets"`     // 交易发起者
	ToWallets      []*TradeWalletItem `json:"to_wallets"`       // 交易接收者
	TradeScene     basic.TradeScene   `json:"trade_scene"`      // 交易场景
	Comment        string             `json:"comment"`          // 交易备注
}

type TradeWalletItem struct {
	WalletId   int64            `json:"wallet_id"` // 交易接收者ID
	Amount     int64            `json:"amount"`    // 接收金额
	ChangeType basic.ChangeType `json:"change_type"`
	Comment    string           `json:"comment"` // 交易备注
}
