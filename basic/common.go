package basic

import (
	"fmt"

	"github.com/pkg/errors"
)

type TransferScene int     //转移场景
type ChangeType int        //变更类型
type RecordStatus int      //转移状态
type RecordType int        //转移类型 增加物品或减少物品
type StateStatus int       //转移状态
type ItemType int          //物品类型
type OfficialBagType int64 //官方背包类型

const (
	DefaultOfficialBagStep = 100000000    //官方背包类型步长 默认1亿
	DefaultOfficialBagMin  = 1            //官方背包最小值 默认1
	DefaultOfficialBagMax  = 100000000000 //官方背包最大值 默认1000亿，即1000个官方背包
	DefaultStateSplitNum   = 1            //转移状态分表数量 默认1单表
	DefaultRecordSplitNum  = 1            //转移记录分表数量 默认1单表
	DefaultBagSplitNum     = 1            //背包分表数量 默认1单表
)

var (
	officialBagStep int64 //官方背包类型步长
	officialBagMin  int64 //官方背包最小值
	officialBagMax  int64 //官方背包最大值

	stateSplitNum  int64 //转移状态分表数量
	recordSplitNum int64 //转移记录分表数量
	bagSplitNum    int64 //背包分表数量
	dbNum          int64 //数据库数量
)

const (
	RecordStatusNormal        RecordStatus = 1 //正常
	RecordStatusRollback      RecordStatus = 2 //回滚
	RecordStatusEmptyRollback RecordStatus = 3 //空回滚
)

const (
	RecordTypeAdd    RecordType = 1 //增加
	RecordTypeDeduct RecordType = 2 //减少
)

const (
	StateStatusDoing         StateStatus = 1 //转移中
	StateStatusRollbackDoing StateStatus = 2 //回滚中
	StateStatusHalfSuccess   StateStatus = 3 //半成功
	StateStatusSuccess       StateStatus = 4 //转移成功
	StateStatusRollbackDone  StateStatus = 5 //回滚完成
)

func initOfficialBag(officialBagStepVal, officialBagMinVal, officialBagMaxVal int64) error {
	if officialBagMaxVal < officialBagStepVal {
		return errors.New("official bag max is less than official bag step")
	}

	if officialBagStepVal <= 0 {
		officialBagStep = DefaultOfficialBagStep
	} else {
		officialBagStep = officialBagStepVal
	}

	if officialBagMinVal <= 0 {
		officialBagMin = DefaultOfficialBagMin
	} else {
		officialBagMin = officialBagMinVal
	}

	if officialBagMaxVal <= 0 {
		officialBagMax = DefaultOfficialBagMax
	} else {
		officialBagMax = officialBagMaxVal
	}
	return nil
}

func initStateSplitNum(num int64) {
	stateSplitNum = num
}

func initRecordSplitNum(num int64) {
	recordSplitNum = num
}

func initBagSplitNum(num int64) {
	bagSplitNum = num
}

func IsOfficialBag(bagId int64) bool {
	return bagId >= officialBagMin && bagId <= officialBagMax
}

func GetRemain(bagId int64) int64 {
	return bagId % officialBagStep
}

func GetMixOfficialBagId(officialBagId, remain int64) int64 {
	if remain == 0 {
		return officialBagId
	}
	return officialBagId - officialBagStep + remain
}

func CheckTransferOfficialBag(bagId int64) bool {
	return bagId%officialBagStep == 0
}

func GetStateTableSuffix(fromBagId int64) string {
	if stateSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", fromBagId%stateSplitNum)
}

func GetRecordTableSuffix(bagId int64) string {
	if recordSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", bagId%recordSplitNum)
}

func GetBagTableSuffix(bagId int64) string {
	if bagSplitNum <= 1 {
		return ""
	}
	return fmt.Sprintf("_%d", bagId%bagSplitNum)
}

func GetStateTableSplitNum() int64 {
	return stateSplitNum
}

func GetDBNum() int64 {
	return dbNum
}
