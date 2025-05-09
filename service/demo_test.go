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

	OfficialAccountTypeBank basic.OfficialAccountType = 100000000

	OfficialAccountTypeFee basic.OfficialAccountType = 100000000000

	TransferSceneBuyGoods        basic.TransferScene = 1
	ChangeTypeSpend              basic.ChangeType    = 1
	ChangeTypeSellGoodsIncome    basic.ChangeType    = 2
	ChangeTypeSellGoodsCopyright basic.ChangeType    = 3
)

func Init(t *testing.T) {
	dsn1 := "root:ERcxF3&72#32q@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	dsn2 := "root:ERcxF3&72#32q@tcp(127.0.0.1:3306)/test2?charset=utf8mb4&parseTime=True&loc=Local"
	// 连接数据库
	db1, err := gorm.Open(mysql.Open(dsn1), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db2, err := gorm.Open(mysql.Open(dsn2), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = basic.InitWithConf(&basic.TransferConf{
		DBs:             []*gorm.DB{db1, db2},
		StateSplitNum:   3,
		RecordSplitNum:  3,
		AccountSplitNum: 3,
	})
	if err != nil {
		t.Fatalf("failed to init conf: %v", err)
	}
}

func TestTransfer(t *testing.T) {
	Init(t)
	ctx := context.Background()
	accountIdOne, accountIdTwo := int64(100000000001), int64(100000000002)
	err := Transfer(ctx, &model.TransferReq{
		FromAccounts: []*model.TransferItem{
			{
				AccountId:  int64(OfficialAccountTypeBank),
				Amount:     100,
				ChangeType: ChangeTypeSpend,
				Comment:    "transfer deduct",
				ItemType:   ItemTypeGold,
			},
		},
		TransferId:     1,
		UseHalfSuccess: true,
		ToAccounts: []*model.TransferItem{
			{
				AccountId:  accountIdOne,
				Amount:     90,
				ChangeType: ChangeTypeSellGoodsIncome,
				Comment:    "transfer sell goods income",
				ItemType:   ItemTypeGold,
			},
			{
				AccountId:  accountIdTwo,
				Amount:     10,
				ChangeType: ChangeTypeSellGoodsCopyright,
				Comment:    "transfer sell goods copyright",
				ItemType:   ItemTypeGold,
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
