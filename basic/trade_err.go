package basic

import (
	"errors"
	"fmt"
)

type TradeErrCode int

const (
	ParamsErrCode             TradeErrCode = 1
	AlreadyRolledBackErrCode  TradeErrCode = 2
	StateMutationErrCode      TradeErrCode = 3
	InsufficientAmountErrCode TradeErrCode = 4
	DBFailedErrCode           TradeErrCode = 5
)

var (
	ParamsErr             = New(ParamsErrCode, "[coin-trade] params error")
	AlreadyRolledBackErr  = New(AlreadyRolledBackErrCode, "[coin-trade] already rolled back")
	StateMutationErr      = New(StateMutationErrCode, "[coin-trade] state mutation")
	InsufficientAmountErr = New(InsufficientAmountErrCode, "[coin-trade] insufficient amount")
	DBFailedErr           = New(DBFailedErrCode, "[coin-trade] db failed")
)

type TradeErr struct {
	Code TradeErrCode
	Msg  string
}

func (e *TradeErr) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

func New(code TradeErrCode, msg string) error {
	return &TradeErr{
		Code: code,
		Msg:  msg,
	}
}

func NewWithErr(code TradeErrCode, err error) error {
	return &TradeErr{
		Code: code,
		Msg:  err.Error(),
	}
}

func NewDBFailed(err error) error {
	return New(DBFailedErrCode, err.Error())
}

func NewParamsError(err error) error {
	return New(ParamsErrCode, err.Error())
}

func Is(err, target error) bool {
	if errors.Is(err, target) {
		return true
	}
	var e *TradeErr
	if errors.As(err, &e) {
		if e.Code == target.(*TradeErr).Code {
			return true
		}
	}
	return false
}
