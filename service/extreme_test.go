package service

import (
	"context"
	"github.com/zjn-zjn/coin-trade/basic"
	"github.com/zjn-zjn/coin-trade/model"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

const (
	concurrentNum = 100
	tradeNum      = 10000
	rollbackNum   = 1000
	inspectionNum = 500
	walletIdBase  = int64(100000000001)
	walletId2Base = int64(100000000002)
)

func TestExtreme(t *testing.T) {
	Init(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	random := rand.New(rand.NewPCG(0, uint64(time.Now().UnixNano())))
	start := time.Now().UnixMilli()
	for c := 0; c < concurrentNum; c++ {
		wg.Add(1)
		go func() {
			for i := 0; i < tradeNum; i++ {
				req := &model.CoinTradeReq{
					FromWalletId:   []int64{int64(OfficialWalletTypeBankWallet), int64(OfficialWalletTypeFee)}[random.IntN(2)],
					FromAmount:     100,
					TradeId:        int64(c*tradeNum + i + 1),
					CoinType:       CoinTypeGold,
					UseHalfSuccess: []bool{true, false}[random.IntN(2)],
					Inverse:        []bool{true, false}[random.IntN(2)],
					TradeScene:     TradeSceneBuyGoods,
					Comment:        "trade goods",
					ToWallets: []*model.TradeWalletItem{
						{
							WalletId: walletIdBase,
							Amount:   90,
							AddType:  AddTypeSellGoodsIncome,
							Comment:  "trade sell goods income",
						},
						{
							WalletId: walletId2Base,
							Amount:   10,
							AddType:  AddTypeSellGoodsIncome,
							Comment:  "trade sell goods copyright",
						},
					},
				}
				err := CoinTrade(ctx, req)
				if err != nil {
					//t.Logf("failed to trade trade_id %d: %v", req.TradeId, err)
				}
				//time.Sleep(1 * time.Millisecond)
			}
			t.Logf("c worker %d done time:%d", c, time.Now().UnixMilli())
			wg.Done()
		}()
	}
	for c := 0; c < concurrentNum; c++ {
		wg.Add(1)
		go func() {
			for i := 0; i < rollbackNum; i++ {
				req := &model.RollbackTradeReq{
					TradeId:    int64(c*tradeNum + i + 1),
					TradeScene: TradeSceneBuyGoods,
				}
				for {
					err := RollbackTrade(ctx, req)
					if err == nil {
						break
					} else {
						//t.Logf("failed to rollback trade trade_id %d: %v", req.TradeId, err)
					}
				}
				time.Sleep(500 * time.Millisecond)
			}
			t.Logf("rollback done time:%d", time.Now().UnixMilli())
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		for i := 0; i < inspectionNum; i++ {
			err := CoinTradeInspection(ctx, time.Now().UnixMilli())
			if err != nil {
				//t.Errorf("failed to inspection trade: %v", err)
			}
			time.Sleep(500 * time.Millisecond)
		}
		t.Logf("inspection done")
		wg.Done()
	}()
	wg.Wait()
	//最后推进一把
	time.Sleep(5 * time.Second)
	t.Logf("final inspection")
	err := CoinTradeInspection(ctx, time.Now().UnixMilli())
	if err != nil {
		t.Logf("failed to inspection trade final: %v", err)
	}
	t.Logf("all done start:%d end:%d", start, time.Now().UnixMilli())
}

func TestFindBadCase(t *testing.T) {
	Init(t)
	ctx := context.Background()
	step := 10000
	minId := 0
	maxId := concurrentNum * tradeNum

	for {
		stateList, err := findTradeStateWithLimit(ctx, int64(minId), int64(minId+step))
		if err != nil {
			t.Fatalf("failed to find trade state: %v", err)
		}
		for _, state := range stateList {
			if state.Status == basic.TradeStateStatusSuccess {
				//对账
				records, err := getTradeRecordsByTradeId(ctx, state.TradeId) //ddl里没有trade_id这个索引，需要的可以加
				if err != nil {
					t.Logf("failed to get trade records: %v", err)
					continue
				}
				var fromAmount, toAmount int64
				for _, record := range records {
					if record.TradeStatus != basic.TradeRecordStatusNormal {
						t.Logf("err trade %d record %d status %d %v", state.TradeId, record.ID, record.TradeStatus, state.Inverse)
						continue
					}
					if record.TradeType == basic.TradeTypeDeduct {
						fromAmount += record.Amount
					} else {
						toAmount += record.Amount
					}
				}
				if fromAmount != toAmount {
					t.Errorf("err trade %d amount %d %d %v", state.TradeId, fromAmount, toAmount, state.Inverse)
					continue
				}
				continue
			}
			if state.Status == basic.TradeStateStatusRollbackDone {
				//对账
				records, err := getTradeRecordsByTradeId(ctx, state.TradeId)
				if err != nil {
					t.Logf("failed to get trade records: %v", err)
					continue
				}
				var fromAmount, toAmount int64
				for _, record := range records {
					if record.TradeStatus == basic.TradeRecordStatusNormal {
						t.Logf("err trade %d record %d status %d %v", state.TradeId, record.ID, record.TradeStatus, state.Inverse)
						continue
					}
					if record.TradeType == basic.TradeTypeDeduct {
						fromAmount += record.Amount
					} else {
						toAmount += record.Amount
					}
				}
				if fromAmount != toAmount {
					t.Logf("err trade %d amount %d %d %v", state.TradeId, fromAmount, toAmount, state.Inverse)
					continue
				}
				continue
			}
			t.Logf("err trade %d status %d %v", state.TradeId, state.Status, state.Inverse)
		}
		if minId >= maxId {
			break
		}
		minId += step
	}
}

func findTradeStateWithLimit(ctx context.Context, idMin, idMax int64) ([]*model.TradeState, error) {
	var stateList []*model.TradeState
	err := basic.GetCoinTradeWriteDB(ctx).Table("trade_state").Where("trade_id >= ? and trade_id <= ?", idMin, idMax).Find(&stateList).Error
	if err != nil {
		return nil, basic.NewDBFailed(err)
	}
	return stateList, nil
}

func getTradeRecordsByTradeId(ctx context.Context, tradeId int64) ([]*model.TradeRecord, error) {
	var records []*model.TradeRecord
	err := basic.GetCoinTradeWriteDB(ctx).Table("trade_record").Where("trade_id = ?", tradeId).Find(&records).Error
	if err != nil {
		return nil, basic.NewDBFailed(err)
	}
	return records, nil
}
