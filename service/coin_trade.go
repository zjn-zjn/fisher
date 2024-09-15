package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/zjn-zjn/coin-trade/basic"
	"github.com/zjn-zjn/coin-trade/dao"
	"github.com/zjn-zjn/coin-trade/model"
)

const (
	walletStateUniqueKey = "%d_%d"
)

// CoinTrade 虚拟币交易
func CoinTrade(ctx context.Context, req *model.CoinTradeReq) error {
	// check args
	err := checkArgs(req)
	if err != nil {
		return basic.NewParamsError(err)
	}
	//处理官方账户,为避免官方账户的热点问题,官方账户采用区间制
	handleOfficialWallet(req)
	//创建交易事务
	state, err := dao.GetOrCreateTradeState(ctx, req)
	if err != nil {
		return err
	}
	if state.Status == basic.TradeStateStatusSuccess || state.Status == basic.TradeStateStatusHalfSuccess {
		//当前事务已成功，为保证幂等性，直接返回成功
		//这里就不怕因网络问题导致的，交易在当前已经成功，但已收到回滚消息，但交易的重试进行到这里，直接返回成功导致的业务异常？
		//不怕，因为rollback一定是发生于正常执行之后，如果已经发生了rollback,那么之前的交易一定是失效的，因此返回交易成功也不会有问题
		//举例：A调用B，A调用超时，B实际已成功，A一次重试调用B，由于发生失败，A于是决定回滚
		//但由于网络问题，到了B这里，可能是重试先到，rollback后到，此时B的重试依然返回了当前交易已成功给到A，无论B的这次返回是否触达A，A都不会将交易继续进行下去
		return nil
	}

	if state.Status == basic.TradeStateStatusRollbackDoing || state.Status == basic.TradeStateStatusRollbackDone {
		//当前事务正在回滚或已回滚，直接返回失败
		return basic.AlreadyRolledBackErr
	}

	if state.Status != basic.TradeStateStatusDoing {
		//当前非doing 直接报错
		return basic.StateMutationErr
	}
	deductionTx, increaseTxs, err := coinTradeSequences(req)
	if err != nil {
		return err
	}

	err = dao.TxWrapper(ctx, state, deductionTx, increaseTxs, req.UseHalfSuccess)
	if err != nil {
		return err
	}
	return nil
}

// 离散官方钱包
func handleOfficialWallet(req *model.CoinTradeReq) {
	walletId := findFirstWalletId(req)
	if walletId == nil {
		//无对标的钱包ID，为官方账户之间流转，直接采用原始值
		return
	}
	musk := basic.GetRemain(*walletId)
	//修改官方账户
	if basic.IsOfficialWallet(req.FromWalletId) {
		req.FromWalletId = basic.GetMixOfficialWalletId(req.FromWalletId, musk)
	}
	for _, v := range req.ToWallets {
		v := v
		if basic.IsOfficialWallet(v.WalletId) {
			v.WalletId = basic.GetMixOfficialWalletId(v.WalletId, musk)
		}
	}
}

func findFirstWalletId(req *model.CoinTradeReq) *int64 {
	if !basic.IsOfficialWallet(req.FromWalletId) {
		return &req.FromWalletId
	}
	for _, v := range req.ToWallets {
		walletId := v.WalletId
		if !basic.IsOfficialWallet(walletId) {
			return &walletId
		}
	}
	return nil
}

func checkArgs(req *model.CoinTradeReq) error {
	if req.FromWalletId <= 0 {
		return errors.New("from wallet illegal")
	}
	if basic.IsOfficialWallet(req.FromWalletId) {
		//官方钱包的传入必须用枚举，由交易系统离散
		if !basic.CheckTradeOfficialWallet(req.FromWalletId) {
			return fmt.Errorf("offical wallet illegal wallet:%d", req.FromWalletId)
		}
	}
	if req.FromAmount <= 0 {
		return errors.New("from amount illegal")
	}
	if req.TradeId <= 0 {
		return errors.New("tradeId illegal")
	}
	if req.CoinType <= 0 {
		return errors.New("coin type illegal")
	}
	if req.TradeScene <= 0 {
		return errors.New("trade scene illegal")
	}
	if len(req.ToWallets) <= 0 {
		return errors.New("to wallets illegal")
	}
	var totalAmount int64
	uniqueCheck := make(map[string]bool)
	fromWalletUniqueKey := fmt.Sprintf(walletStateUniqueKey, req.FromWalletId, req.TradeScene)
	uniqueCheck[fromWalletUniqueKey] = true
	for _, toWalletInfo := range req.ToWallets {
		if toWalletInfo.WalletId <= 0 {
			return fmt.Errorf("[coin-trade] to wallet illegal to_wallet:%d", toWalletInfo.WalletId)
		}
		if basic.IsOfficialWallet(toWalletInfo.WalletId) {
			//官方钱包的传入必须用枚举，由交易系统离散
			if !basic.CheckTradeOfficialWallet(toWalletInfo.WalletId) {
				return fmt.Errorf("[coin-trade] offical wallet illegal to_wallet:%d", toWalletInfo.WalletId)
			}
		}
		//非官方钱包不允许操作为0的金额
		if toWalletInfo.Amount <= 0 && !basic.IsOfficialWallet(toWalletInfo.WalletId) {
			return fmt.Errorf("[coin-trade] to amount illegal to_amount:%d", toWalletInfo.Amount)
		}
		totalAmount = toWalletInfo.Amount + totalAmount
		//用于校验钱包交易场景是否重复，如：钱包A1向A2和A3转账，A1 A2 A3的扣钱和加钱场景不应有重叠(比如A2 A3不能都是版权结算)
		toWalletUniqueKey := fmt.Sprintf(walletStateUniqueKey, toWalletInfo.WalletId, toWalletInfo.AddType)
		if _, ok := uniqueCheck[toWalletUniqueKey]; ok {
			return fmt.Errorf("[coin-trade] wallet duplicate change type wallet:%d", toWalletInfo.WalletId)
		}
		uniqueCheck[toWalletUniqueKey] = true
	}
	if totalAmount != req.FromAmount {
		return errors.New("[coin-trade] unable to balance accounts")
	}
	return nil
}

func coinTradeSequences(req *model.CoinTradeReq) (*dao.TradeTxItem, []*dao.TradeTxItem, error) {
	//扣除金额
	deductionTx := &dao.TradeTxItem{
		Exec: func(ctx context.Context) error {
			err := dao.DeductionWallet(ctx, req.FromWalletId, req.TradeId, req.FromAmount, req.CoinType, req.TradeScene, basic.TradeRecordStatusNormal, basic.ChangeType(req.TradeScene), req.Comment)
			if err != nil {
				return err
			}
			return nil
		},
		Rollback: func(ctx context.Context) error {
			//回滚增加金额
			err := dao.IncreaseWallet(ctx, req.FromWalletId, req.TradeId, req.FromAmount, req.CoinType, req.TradeScene, basic.TradeRecordStatusRollback, basic.ChangeType(req.TradeScene), req.Comment)
			if err != nil {
				return err
			}
			return nil
		},
	}
	var increaseTxs []*dao.TradeTxItem
	//增加金额
	for _, toWalletInfo := range req.ToWallets {
		toWalletInfo := toWalletInfo
		increaseTxs = append(increaseTxs, &dao.TradeTxItem{
			Exec: func(ctx context.Context) error {
				// 增加金额
				comment := req.Comment
				if toWalletInfo.Comment != "" {
					comment = toWalletInfo.Comment
				}
				err := dao.IncreaseWallet(ctx, toWalletInfo.WalletId, req.TradeId, toWalletInfo.Amount, req.CoinType, req.TradeScene, basic.TradeRecordStatusNormal, basic.ChangeType(toWalletInfo.AddType), comment)
				if err != nil {
					return err
				}
				return nil
			},
			Rollback: func(ctx context.Context) error {
				//回滚扣除金额
				err := dao.DeductionWallet(ctx, toWalletInfo.WalletId, req.TradeId, toWalletInfo.Amount, req.CoinType, req.TradeScene, basic.TradeRecordStatusRollback, basic.ChangeType(toWalletInfo.AddType), req.Comment)
				if err != nil {
					return err
				}
				return nil
			},
		})
	}
	return deductionTx, increaseTxs, nil
}
