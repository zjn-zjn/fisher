package model

import "github.com/zjn-zjn/coin-trade/basic"

type CoinTradeReq struct {
	FromWalletId int64                    `json:"from_wallet_id"` // 交易发起者
	FromAmount   int64                    `json:"from_amount"`    // 交易金额
	TradeId      int64                    `json:"trade_id"`       // 交易ID
	CoinType     basic.CoinType           `json:"coin_type"`      // 交易币种
	ToWallets    []*CoinTradeToWalletItem `json:"to_wallets"`     // 交易接收者
	TradeScene   basic.TradeScene         `json:"trade_scene"`    // 交易场景
	Comment      string                   `json:"comment"`        // 交易备注
}

type CoinTradeToWalletItem struct {
	WalletId int64         `json:"wallet_id"` // 交易接收者ID
	Amount   int64         `json:"amount"`    // 接收金额
	AddType  basic.AddType `json:"add_type"`  // 交易类型
	Comment  string        `json:"comment"`   // 交易备注
}
