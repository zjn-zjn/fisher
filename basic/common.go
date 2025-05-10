package basic

import (
	"fmt"

	"github.com/pkg/errors"
)

type TransferScene int         //转移场景
type ChangeType int            //变更类型
type RecordStatus int          //记录状态
type TransferType int          //转移类型 增加物品或减少物品
type StateStatus int           //转移状态
type ItemType int              //物品类型
type OfficialAccountType int64 //官方账户类型

const (
	DefaultOfficialAccountStep = 10000000    //官方账户类型步长 默认1千万
	DefaultOfficialAccountMin  = 1           //官方账户最小值 默认1
	DefaultOfficialAccountMax  = 10000000000 //官方账户最大值 默认100亿，即1000个官方账户
	DefaultStateSplitNum       = 1           //转移状态分表数量 默认1单表
	DefaultRecordSplitNum      = 1           //转移记录分表数量 默认1单表
	DefaultAccountSplitNum     = 1           //账户分表数量 默认1单表
)

var (
	officialAccountStep int64 //官方账户类型步长
	officialAccountMin  int64 //官方账户最小值
	officialAccountMax  int64 //官方账户最大值

	stateSplitNum   int64 //转移状态分表数量
	recordSplitNum  int64 //转移记录分表数量
	accountSplitNum int64 //账户分表数量
	dbNum           int64 //数据库数量
)

const (
	RecordStatusNormal        RecordStatus = 1 //正常
	RecordStatusRollback      RecordStatus = 2 //回滚
	RecordStatusEmptyRollback RecordStatus = 3 //空回滚
)

const (
	RecordTypeAdd    TransferType = 1 //增加
	RecordTypeDeduct TransferType = 2 //减少
)

const (
	StateStatusDoing         StateStatus = 1 //转移中
	StateStatusRollbackDoing StateStatus = 2 //回滚中
	StateStatusHalfSuccess   StateStatus = 3 //半成功
	StateStatusSuccess       StateStatus = 4 //转移成功
	StateStatusRollbackDone  StateStatus = 5 //回滚完成
)

func initOfficialAccount(officialAccountStepVal, officialAccountMinVal, officialAccountMaxVal int64) error {
	if officialAccountMaxVal < officialAccountStepVal {
		return errors.New("official account max is less than official account step")
	}

	if officialAccountStepVal <= 0 {
		officialAccountStep = DefaultOfficialAccountStep
	} else {
		officialAccountStep = officialAccountStepVal
	}

	if officialAccountMinVal <= 0 {
		officialAccountMin = DefaultOfficialAccountMin
	} else {
		officialAccountMin = officialAccountMinVal
	}

	if officialAccountMaxVal <= 0 {
		officialAccountMax = DefaultOfficialAccountMax
	} else {
		officialAccountMax = officialAccountMaxVal
	}
	return nil
}

func initStateSplitNum(num int64) {
	stateSplitNum = num
}

func initRecordSplitNum(num int64) {
	recordSplitNum = num
}

func initAccountSplitNum(num int64) {
	accountSplitNum = num
}

func IsOfficialAccount(accountId int64) bool {
	return accountId >= officialAccountMin && accountId <= officialAccountMax
}

func GetRemain(accountId int64) int64 {
	return accountId % officialAccountStep
}

func GetMixOfficialAccountId(officialAccountId, remain int64) int64 {
	if remain == 0 {
		return officialAccountId
	}
	return officialAccountId - officialAccountStep + remain
}

func CheckTransferOfficialAccount(accountId int64) bool {
	return accountId%officialAccountStep == 0
}

func GetStateTableSuffix(fromAccountId int64) string {
	if stateSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", fromAccountId%stateSplitNum)
}

func GetRecordTableSuffix(accountId int64) string {
	if recordSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", accountId%recordSplitNum)
}

func GetAccountTableSuffix(accountId int64) string {
	if accountSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", accountId%accountSplitNum)
}

func GetStateTableSplitNum() int64 {
	return stateSplitNum
}

func GetDBNum() int64 {
	return dbNum
}
