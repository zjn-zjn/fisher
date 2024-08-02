package basic

import (
	"fmt"

	"github.com/pkg/errors"
)

type TradeScene int           //交易场景
type AddType int              //收款方虚拟币增加类型
type ChangeType int           //付款方的traceScene和收款方的addType 合并字段
type TradeRecordStatus int    //交易状态
type TradeType int            //交易类型 增加虚拟币或减少虚拟币
type TradeStateStatus int     //交易状态
type CoinType int             //虚拟币类型
type OfficialWalletType int64 //官方钱包类型

const (
	DefaultOfficialWalletStep  = 100000000    //官方钱包类型步长 默认1亿
	DefaultOfficialWalletMax   = 100000000000 //官方钱包最大值 默认1000亿
	DefaultTradeStateSplitNum  = 1            //交易状态分表数量 默认1单表
	DefaultTradeRecordSplitNum = 1            //交易记录分表数量 默认1单表
	DefaultWalletBagSplitNum   = 1            //钱包分表数量 默认1单表
)

var (
	officialWalletStep int64 //官方钱包类型步长
	officialWalletMax  int64 //官方钱包最大值

	tradeStateSplitNum  int64 //交易状态分表数量
	tradeRecordSplitNum int64 //交易记录分表数量
	walletBagSplitNum   int64 //钱包分表数量
)

const (
	TradeRecordStatusNormal        TradeRecordStatus = 1 //正常
	TradeRecordStatusRollback      TradeRecordStatus = 2 //回滚
	TradeRecordStatusEmptyRollback TradeRecordStatus = 3 //空回滚
)

const (
	TradeTypeAdd    TradeType = 1 //增加
	TradeTypeDeduct TradeType = 2 //减少
)

const (
	TradeStateStatusDoing         TradeStateStatus = 1 //交易中
	TradeStateStatusRollbackDoing TradeStateStatus = 2 //回滚中
	TradeStateStatusHalfSuccess   TradeStateStatus = 3 //TODO 部分成功(可以持续推进成成功的) 暂未实现
	TradeStateStatusSuccess       TradeStateStatus = 4 //交易成功
	TradeStateStatusRollbackDone  TradeStateStatus = 5 //回滚完成
)

func InitOfficialWallet(officialWalletStepVal, officialWalletMaxVal int64) error {
	if officialWalletMaxVal < officialWalletStepVal {
		return errors.New("official wallet max is less than official wallet step")
	}
	if officialWalletStepVal <= 0 {
		officialWalletStep = DefaultOfficialWalletStep
	} else {
		officialWalletStep = officialWalletStepVal
	}
	if officialWalletMaxVal <= 0 {
		officialWalletMax = DefaultOfficialWalletMax
	} else {
		officialWalletMax = officialWalletMaxVal
	}
	return nil
}

func InitTradeStateSplitNum(num int64) {
	tradeStateSplitNum = num
}

func InitTradeRecordSplitNum(num int64) {
	tradeRecordSplitNum = num
}

func InitWalletBagSplitNum(num int64) {
	walletBagSplitNum = num
}

func IsOfficialWallet(walletId int64) bool {
	return walletId <= officialWalletMax && walletId > 0
}

func GetRemain(walletId int64) int64 {
	return walletId % officialWalletStep
}

func GetMixOfficialWalletId(officialWalletId, remain int64) int64 {
	if remain == 0 {
		return officialWalletId
	}
	return officialWalletId - officialWalletStep + remain
}

func CheckTradeOfficialWallet(walletId int64) bool {
	return walletId%officialWalletStep == 0
}

func GetTradeStateTableSuffix(fromWalletId int64) string {
	if tradeStateSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", fromWalletId%tradeStateSplitNum)
}

func GetTradeRecordTableSuffix(walletId int64) string {
	if tradeRecordSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", walletId%tradeRecordSplitNum)
}

func GetWalletBagTableSuffix(walletId int64) string {
	if walletBagSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", walletId%walletBagSplitNum)
}

func GetTradeStateTableSplitNum() int64 {
	return tradeStateSplitNum
}
