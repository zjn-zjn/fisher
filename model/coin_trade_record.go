package model

import "github.com/zjn-zjn/coin-trade/basic"

const (
	TradeRecordTablePrefix = "trade_record"
)

type TradeRecord struct {
	ID          int64                   `json:"id" gorm:"column:id;"`                                     // 主键
	WalletId    int64                   `json:"wallet_id" gorm:"column:wallet_id;"`                       // 钱包ID
	TradeId     int64                   `json:"trade_id" gorm:"column:trade_id;"`                         // 交易ID
	TradeScene  basic.TradeScene        `json:"trade_scene" gorm:"column:trade_scene;"`                   // 交易场景
	TradeType   basic.TradeType         `json:"trade_type" gorm:"column:trade_type;"`                     // 交易类型
	TradeStatus basic.TradeRecordStatus `json:"trade_status" gorm:"column:trade_status;"`                 // 交易状态
	Amount      int64                   `json:"amount" gorm:"column:amount;"`                             // 交易金额
	CoinType    basic.CoinType          `json:"coin_type" gorm:"column:coin_type;"`                       // 交易币种
	ChangeType  basic.ChangeType        `json:"change_type" gorm:"column:change_type;"`                   // 交易变化类型
	Comment     string                  `json:"comment" gorm:"column:comment;"`                           // 交易备注
	CreatedAt   int64                   `json:"created_at" gorm:"column:created_at;autoCreateTime:milli"` // 创建时间
	UpdatedAt   int64                   `json:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli"` // 创建时间
}

func GetTradeRecordTableName(walletId int64) string {
	return TradeRecordTablePrefix + basic.GetTradeRecordTableSuffix(walletId)
}
