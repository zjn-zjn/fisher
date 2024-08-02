package model

import "github.com/zjn-zjn/coin-trade/basic"

const (
	WalletBagTablePrefix = "wallet_bag"
)

type WalletBag struct {
	ID        int64          `json:"id" gorm:"column:id;"`                                     // 主键
	WalletId  int64          `json:"wallet_id" gorm:"column:wallet_id;"`                       // 钱包ID
	CoinType  basic.CoinType `json:"coin_type" gorm:"column:coin_type;"`                       // 交易币种
	Amount    int64          `json:"amount" gorm:"column:amount;"`                             // 余额
	CreatedAt int64          `json:"created_at" gorm:"column:created_at;autoCreateTime:milli"` // 创建时间
	UpdatedAt int64          `json:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli"` // 创建时间
}

func GetWalletBagTableName(walletId int64) string {
	return WalletBagTablePrefix + basic.GetWalletBagTableSuffix(walletId)
}
