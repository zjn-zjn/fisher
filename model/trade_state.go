package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/zjn-zjn/coin-trade/basic"
)

const (
	TradeStateTablePrefix = "trade_state"
)

type TradeState struct {
	ID           int64                  `json:"id" gorm:"column:id;"`                                     // 主键
	TradeId      int64                  `json:"trade_id" gorm:"column:trade_id;"`                         // 交易ID
	TradeScene   basic.TradeScene       `json:"trade_scene" gorm:"column:trade_scene;"`                   // 交易场景
	FromWalletId int64                  `json:"from_wallet_id" gorm:"column:from_wallet_id;"`             // 扣款钱包ID
	FromAmount   int64                  `json:"from_amount" gorm:"column:from_amount;"`                   // 交易金额
	ToWallets    ToWalletList           `json:"to_wallets" gorm:"column:to_wallets;"`                     // 收款钱包信息列表
	CoinType     basic.CoinType         `json:"coin_type" gorm:"column:coin_type;"`                       // 虚拟币类型
	Status       basic.TradeStateStatus `json:"status" gorm:"column:status;"`                             // 交易状态 1-进行中 2-回滚中 3-半成功 4-成功 5-已回滚
	Comment      string                 `json:"comment" gorm:"column:comment;"`                           // 交易备注
	CreatedAt    int64                  `json:"created_at" gorm:"column:created_at;autoCreateTime:milli"` // 创建时间
	UpdatedAt    int64                  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli"` // 创建时间
}

type ToWalletList []*CoinTradeToWalletItem

func GetTradeStateTableName(tradeId int64) string {
	return TradeStateTablePrefix + basic.GetTradeStateTableSuffix(tradeId)
}

func AssembleTradeState(toWallets []*CoinTradeToWalletItem, fromWalletId, tradeId, fromAmount int64, tradeScene basic.TradeScene, status basic.TradeStateStatus, coinType basic.CoinType, comment string) *TradeState {
	md := &TradeState{
		TradeId:      tradeId,
		TradeScene:   tradeScene,
		FromWalletId: fromWalletId,
		FromAmount:   fromAmount,
		CoinType:     coinType,
		Status:       status,
		Comment:      comment,
	}
	if toWallets == nil {
		md.ToWallets = make(ToWalletList, 0)
	} else {
		md.ToWallets = toWallets
	}
	return md
}

func (m *ToWalletList) Scan(val interface{}) error {
	s := val.([]uint8)
	var toWallets ToWalletList
	err := json.Unmarshal(s, &toWallets)
	if err != nil {
		return err
	}
	*m = toWallets
	return nil
}

func (m ToWalletList) Value() (driver.Value, error) {
	result, _ := json.Marshal(m)
	return string(result), nil
}
