package basic

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type TransferConf struct {
	DBs                 []*gorm.DB `json:"-"`
	StateSplitNum       int64      `json:"state_split_num"`       //转移状态分表数量 按转移ID取模分表
	RecordSplitNum      int64      `json:"record_split_num"`      //转移记录分表数量 按账户ID取模分表
	AccountSplitNum     int64      `json:"account_split_num"`     //账户分表数量 按账户ID取模分表
	OfficialAccountStep int64      `json:"official_account_step"` //官方账户类型步长
	OfficialAccountMin  int64      `json:"official_account_min"`  //官方账户最小值
	OfficialAccountMax  int64      `json:"official_account_max"`  //官方账户最大值
}

// InitWithDefault 使用默认配置初始化
func InitWithDefault(dbs []*gorm.DB) error {
	return InitWithConf(&TransferConf{
		DBs:                 dbs,
		StateSplitNum:       DefaultStateSplitNum,
		RecordSplitNum:      DefaultRecordSplitNum,
		AccountSplitNum:     DefaultAccountSplitNum,
		OfficialAccountStep: DefaultOfficialAccountStep,
		OfficialAccountMin:  DefaultOfficialAccountMin,
		OfficialAccountMax:  DefaultOfficialAccountMax,
	})
}

// InitWithConf 使用配置初始化
func InitWithConf(conf *TransferConf) error {
	if conf == nil {
		return errors.New("conf is nil")
	}
	if len(conf.DBs) == 0 {
		return errors.New("db is nil")
	}
	initItemTransferDB(conf.DBs)
	err := initOfficialAccount(conf.OfficialAccountStep, conf.OfficialAccountMin, conf.OfficialAccountMax)
	if err != nil {
		return err
	}
	initStateSplitNum(conf.StateSplitNum)
	initRecordSplitNum(conf.RecordSplitNum)
	initAccountSplitNum(conf.AccountSplitNum)
	return nil
}
