package basic

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type TransferConf struct {
	DB              *gorm.DB `json:"-"`
	StateSplitNum   int64    `json:"state_split_num"`   //转移状态分表数量 按转移ID取模分表
	RecordSplitNum  int64    `json:"record_split_num"`  //转移记录分表数量 按背包ID取模分表
	BagSplitNum     int64    `json:"bag_split_num"`     //背包分表数量 按背包ID取模分表
	OfficialBagStep int64    `json:"official_bag_step"` //官方背包类型步长
	OfficialBagMax  int64    `json:"official_bag_max"`  //官方背包最大值
}

// InitWithDefault 使用默认配置初始化
func InitWithDefault(db *gorm.DB) error {
	return InitWithConf(&TransferConf{
		DB:              db,
		StateSplitNum:   DefaultStateSplitNum,
		RecordSplitNum:  DefaultRecordSplitNum,
		BagSplitNum:     DefaultBagSplitNum,
		OfficialBagStep: DefaultOfficialBagStep,
		OfficialBagMax:  DefaultOfficialBagMax,
	})
}

// InitWithConf 使用配置初始化
func InitWithConf(conf *TransferConf) error {
	if conf == nil {
		return errors.New("conf is nil")
	}
	if conf.DB == nil {
		return errors.New("db is nil")
	}
	initItemTransferDB(conf.DB)
	err := initOfficialBag(conf.OfficialBagStep, conf.OfficialBagMax)
	if err != nil {
		return err
	}
	initStateSplitNum(conf.StateSplitNum)
	initRecordSplitNum(conf.RecordSplitNum)
	initBagSplitNum(conf.BagSplitNum)
	return nil
}
