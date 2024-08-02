package conf

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/zjn-zjn/coin-trade/basic"
)

type TradeConf struct {
	DB                  *gorm.DB `json:"-"`
	TradeStateSplitNum  int64    `json:"trade_state_split_num"`  //交易状态分表数量 按交易ID取模分表
	TradeRecordSplitNum int64    `json:"trade_record_split_num"` //交易记录分表数量 按钱包ID取模分表
	WalletBagSplitNum   int64    `json:"wallet_bag_split_num"`   //钱包分表数量 按钱包ID取模分表
	OfficialWalletStep  int64    `json:"official_wallet_step"`   //官方钱包类型步长
	OfficialWalletMax   int64    `json:"official_wallet_max"`    //官方钱包最大值
}

func InitWithDefault(db *gorm.DB) error {
	return InitWithConf(&TradeConf{
		DB:                  db,
		TradeStateSplitNum:  basic.DefaultTradeStateSplitNum,
		TradeRecordSplitNum: basic.DefaultTradeRecordSplitNum,
		WalletBagSplitNum:   basic.DefaultWalletBagSplitNum,
		OfficialWalletStep:  basic.DefaultOfficialWalletStep,
		OfficialWalletMax:   basic.DefaultOfficialWalletMax,
	})
}

func InitWithConf(conf *TradeConf) error {
	if conf == nil {
		return errors.New("conf is nil")
	}
	if conf.DB == nil {
		return errors.New("db is nil")
	}
	initCoinTradeDB(conf.DB)
	err := basic.InitOfficialWallet(conf.OfficialWalletStep, conf.OfficialWalletMax)
	if err != nil {
		return err
	}
	basic.InitTradeStateSplitNum(conf.TradeStateSplitNum)
	basic.InitTradeRecordSplitNum(conf.TradeRecordSplitNum)
	basic.InitWalletBagSplitNum(conf.WalletBagSplitNum)
	return nil
}
