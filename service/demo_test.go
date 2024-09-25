package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"
)

const (
	ItemTypeGold basic.ItemType = 1

	OfficialBagTypeBank basic.OfficialBagType = 100000000

	OfficialBagTypeFee basic.OfficialBagType = 200000000

	TransferSceneBuyGoods        basic.TransferScene = 1
	ChangeTypeSpend              basic.ChangeType    = 1
	ChangeTypeSellGoodsIncome    basic.ChangeType    = 2
	ChangeTypeSellGoodsCopyright basic.ChangeType    = 3
)

func Init(t *testing.T) {
	dsn := "root:ERcxF3&72#32q@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = basic.InitWithConf(&basic.TransferConf{
		DBs:            []*gorm.DB{db},
		StateSplitNum:  3,
		RecordSplitNum: 3,
		BagSplitNum:    3,
	})
	if err != nil {
		t.Fatalf("failed to init conf: %v", err)
	}
}

func TestTransfer(t *testing.T) {
	Init(t)
	ctx := context.Background()
	bagIdOne, bagIdTwo := int64(100000000001), int64(100000000002)
	err := Transfer(ctx, &model.TransferReq{
		FromBags: []*model.TransferItem{
			{
				BagId:      int64(OfficialBagTypeBank),
				Amount:     100,
				ChangeType: ChangeTypeSpend,
				Comment:    "transfer deduct",
			},
		},
		TransferId:     1,
		ItemType:       ItemTypeGold,
		UseHalfSuccess: true,
		ToBags: []*model.TransferItem{
			{
				BagId:      bagIdOne,
				Amount:     90,
				ChangeType: ChangeTypeSellGoodsIncome,
				Comment:    "transfer sell goods income",
			},
			{
				BagId:      bagIdTwo,
				Amount:     10,
				ChangeType: ChangeTypeSellGoodsCopyright,
				Comment:    "transfer sell goods copyright",
			},
		},
		TransferScene: TransferSceneBuyGoods,
		Comment:       "transfer goods",
	})
	if err != nil {
		if basic.Is(err, basic.AlreadyRolledBackErr) {
			t.Logf("transfer has been rolled back: %v", err)
		}
		if basic.Is(err, basic.ParamsErr) {
			t.Logf("transfer params error: %v", err)
		}
		if basic.Is(err, basic.DBFailedErr) {
			t.Logf("transfer db failed: %v", err)
		}
		if basic.Is(err, basic.StateMutationErr) {
			t.Logf("transfer state mutation error: %v", err)
		}
		if basic.Is(err, basic.InsufficientAmountErr) {
			t.Logf("transfer insufficient amount: %v", err)
		}
		t.Logf("failed to item transfer: %v", err)
	}
	time.Sleep(time.Second)
}

func TestRollback(t *testing.T) {
	Init(t)
	ctx := context.Background()
	err := Rollback(ctx, &model.RollbackReq{
		TransferId:    1,
		TransferScene: TransferSceneBuyGoods,
	})
	if err != nil {
		t.Fatalf("failed to rollback transfer: %v", err)
	}
}

func TestInspection(t *testing.T) {
	Init(t)
	ctx := context.Background()
	errs := Inspection(ctx, time.Now().UnixMilli())
	if len(errs) > 0 {
		t.Fatalf("failed to item transfer inspection: %v", errs)
	}
}
