package service

import (
	"context"
	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

const (
	concurrentNum = 100
	transferNum   = 10000
	rollbackNum   = 1000
	inspectionNum = 500
	bagIdBase     = int64(100000000001)
	bagId2Base    = int64(100000000002)
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
			for i := 0; i < transferNum; i++ {
				req := &model.TransferReq{
					FromBags: []*model.TransferItem{
						{
							BagId:      []int64{int64(OfficialBagTypeBankBag), int64(OfficialBagTypeFee)}[random.IntN(2)],
							Amount:     100,
							ChangeType: ChangeTypeSpend,
							Comment:    "transfer deduct",
						},
					},
					TransferId:     int64(c*transferNum + i + 1),
					ItemType:       ItemTypeGold,
					UseHalfSuccess: []bool{true, false}[random.IntN(2)],
					TransferScene:  TransferSceneBuyGoods,
					Comment:        "transfer goods",
					ToBags: []*model.TransferItem{
						{
							BagId:      bagIdBase,
							Amount:     90,
							ChangeType: ChangeTypeSellGoodsIncome,
							Comment:    "transfer sell goods income",
						},
						{
							BagId:      bagId2Base,
							Amount:     10,
							ChangeType: ChangeTypeSellGoodsCopyright,
							Comment:    "transfer sell goods copyright",
						},
					},
				}
				err := Transfer(ctx, req)
				if err != nil {
					//t.Logf("failed to transfer transfer_id %d: %v", req.TransferId, err)
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
				req := &model.RollbackReq{
					TransferId:    int64(c*transferNum + i + 1),
					TransferScene: TransferSceneBuyGoods,
				}
				for {
					err := Rollback(ctx, req)
					if err == nil {
						break
					} else {
						//t.Logf("failed to rollback transfer transfer_id %d: %v", req.TransferId, err)
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
			err := Inspection(ctx, time.Now().UnixMilli())
			if err != nil {
				//t.Errorf("failed to inspection transfer: %v", err)
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
	err := Inspection(ctx, time.Now().UnixMilli())
	if err != nil {
		t.Logf("failed to inspection transfer final: %v", err)
	}
	t.Logf("all done start:%d end:%d", start, time.Now().UnixMilli())
}

func TestFindBadCase(t *testing.T) {
	Init(t)
	ctx := context.Background()
	step := 10000
	minId := 0
	maxId := concurrentNum * transferNum

	for {
		stateList, err := findStateWithLimit(ctx, int64(minId), int64(minId+step))
		if err != nil {
			t.Fatalf("failed to find transfer state: %v", err)
		}
		for _, state := range stateList {
			if state.Status == basic.StateStatusSuccess {
				//对账
				records, err := getRecordsByTransferId(ctx, state.TransferId) //ddl里没有transfer_id这个索引，需要的可以加
				if err != nil {
					t.Logf("failed to get transfer records: %v", err)
					continue
				}
				var fromAmount, toAmount int64
				for _, record := range records {
					if record.TransferStatus != basic.RecordStatusNormal {
						t.Logf("err transfer %d record %d status %d", state.TransferId, record.ID, record.TransferStatus)
						continue
					}
					if record.RecordType == basic.RecordTypeDeduct {
						fromAmount += record.Amount
					} else {
						toAmount += record.Amount
					}
				}
				if fromAmount != toAmount {
					t.Errorf("err transfer %d amount %d %d", state.TransferId, fromAmount, toAmount)
					continue
				}
				continue
			}
			if state.Status == basic.StateStatusRollbackDone {
				//对账
				records, err := getRecordsByTransferId(ctx, state.TransferId)
				if err != nil {
					t.Logf("failed to get transfer records: %v", err)
					continue
				}
				var fromAmount, toAmount int64
				for _, record := range records {
					if record.TransferStatus == basic.RecordStatusNormal {
						t.Logf("err transfer %d record %d status %d", state.TransferId, record.ID, record.TransferStatus)
						continue
					}
					if record.RecordType == basic.RecordTypeDeduct {
						fromAmount += record.Amount
					} else {
						toAmount += record.Amount
					}
				}
				if fromAmount != toAmount {
					t.Logf("err transfer %d amount %d %d", state.TransferId, fromAmount, toAmount)
					continue
				}
				continue
			}
			t.Logf("err transfer %d status %d", state.TransferId, state.Status)
		}
		if minId >= maxId {
			break
		}
		minId += step
	}
}

func findStateWithLimit(ctx context.Context, idMin, idMax int64) ([]*model.State, error) {
	var stateList []*model.State
	err := basic.GetWriteDB(ctx).Table("state").Where("transfer_id >= ? and transfer_id <= ?", idMin, idMax).Find(&stateList).Error
	if err != nil {
		return nil, basic.NewDBFailed(err)
	}
	return stateList, nil
}

func getRecordsByTransferId(ctx context.Context, transferId int64) ([]*model.Record, error) {
	var records []*model.Record
	err := basic.GetWriteDB(ctx).Table("record").Where("transfer_id = ?", transferId).Find(&records).Error
	if err != nil {
		return nil, basic.NewDBFailed(err)
	}
	return records, nil
}
