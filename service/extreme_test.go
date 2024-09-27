package service

import (
	"context"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/zjn-zjn/fisher/basic"
	"github.com/zjn-zjn/fisher/model"
)

const (
	concurrentNum = 100
	transferNum   = 10000
	rollbackNum   = 1000
	inspectionNum = 500
	bagIdBase     = int64(100000000001)
	bagIdBaseEnd  = int64(100010000001)
	bagId2Base    = int64(100000000002)
	bagId2BaseEnd = int64(100010000002)
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
				c1 := random.Int64N(1000) + 1
				c2 := random.Int64N(1000) + 1
				req := &model.TransferReq{
					FromBags: []*model.TransferItem{
						{
							BagId:      (random.Int64N(basic.DefaultOfficialBagMax) / basic.DefaultOfficialBagStep * basic.DefaultOfficialBagStep) + basic.DefaultOfficialBagStep,
							Amount:     c1 + c2,
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
							BagId:      bagIdBase + random.Int64N(bagIdBaseEnd-bagIdBase),
							Amount:     c1,
							ChangeType: ChangeTypeSellGoodsIncome,
							Comment:    "transfer sell goods income",
						},
						{
							BagId:      bagId2Base + random.Int64N(bagId2BaseEnd-bagId2Base),
							Amount:     c2,
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
