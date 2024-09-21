package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/zjn-zjn/coin-trade/basic"
	"github.com/zjn-zjn/coin-trade/model"
)

const (
	CoinTypeGold basic.CoinType = 1

	OfficialWalletTypeBankWallet basic.OfficialWalletType = 100000000

	OfficialWalletTypeFee basic.OfficialWalletType = 200000000

	TradeSceneBuyGoods           basic.TradeScene = 1
	ChangeTypeSpend              basic.ChangeType = 1
	ChangeTypeSellGoodsIncome    basic.ChangeType = 2
	ChangeTypeSellGoodsCopyright basic.ChangeType = 3
)

func Init(t *testing.T) {
	dsn := "root:ERcxF3&72#32q@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = basic.InitWithConf(&basic.TradeConf{
		DB:                  db,
		TradeStateSplitNum:  1,
		TradeRecordSplitNum: 1,
		WalletBagSplitNum:   1,
	})
	if err != nil {
		t.Fatalf("failed to init conf: %v", err)
	}
}

func TestCoinTrade(t *testing.T) {
	Init(t)
	ctx := context.Background()
	walletIdOne, walletIdTwo := int64(100000000001), int64(100000000002)
	err := CoinTrade(ctx, &model.CoinTradeReq{
		FromWallets: []*model.TradeWalletItem{
			{
				WalletId:   int64(OfficialWalletTypeBankWallet),
				Amount:     100,
				ChangeType: ChangeTypeSpend,
				Comment:    "trade deduct",
			},
		},
		TradeId:        1,
		CoinType:       CoinTypeGold,
		UseHalfSuccess: true,
		ToWallets: []*model.TradeWalletItem{
			{
				WalletId:   walletIdOne,
				Amount:     90,
				ChangeType: ChangeTypeSellGoodsIncome,
				Comment:    "trade sell goods income",
			},
			{
				WalletId:   walletIdTwo,
				Amount:     10,
				ChangeType: ChangeTypeSellGoodsCopyright,
				Comment:    "trade sell goods copyright",
			},
		},
		TradeScene: TradeSceneBuyGoods,
		Comment:    "trade goods",
	})
	if err != nil {
		if basic.Is(err, basic.AlreadyRolledBackErr) {
			t.Logf("trade has been rolled back: %v", err)
		}
		if basic.Is(err, basic.ParamsErr) {
			t.Logf("trade params error: %v", err)
		}
		if basic.Is(err, basic.DBFailedErr) {
			t.Logf("trade db failed: %v", err)
		}
		if basic.Is(err, basic.StateMutationErr) {
			t.Logf("trade state mutation error: %v", err)
		}
		if basic.Is(err, basic.InsufficientAmountErr) {
			t.Logf("trade insufficient amount: %v", err)
		}
		t.Logf("failed to coin trade: %v", err)
	}
	time.Sleep(time.Second)
}

func TestRollbackTrade(t *testing.T) {
	Init(t)
	ctx := context.Background()
	err := RollbackTrade(ctx, &model.RollbackTradeReq{
		TradeId:    1,
		TradeScene: TradeSceneBuyGoods,
	})
	if err != nil {
		t.Fatalf("failed to rollback trade: %v", err)
	}
}

func TestCoinTradeInspection(t *testing.T) {
	Init(t)
	ctx := context.Background()
	errs := CoinTradeInspection(ctx, time.Now().UnixMilli())
	if len(errs) > 0 {
		t.Fatalf("failed to coin trade inspection: %v", errs)
	}
}
