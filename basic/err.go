package basic

import (
	"errors"
	"fmt"
)

type ErrCode int

const (
	ParamsErrCode             ErrCode = 1
	AlreadyRolledBackErrCode  ErrCode = 2
	StateMutationErrCode      ErrCode = 3
	InsufficientAmountErrCode ErrCode = 4
	DBFailedErrCode           ErrCode = 5
)

var (
	ParamsErr             = New(ParamsErrCode, "[fisher] params error")
	AlreadyRolledBackErr  = New(AlreadyRolledBackErrCode, "[fisher] already rolled back")
	StateMutationErr      = New(StateMutationErrCode, "[fisher] state mutation")
	InsufficientAmountErr = New(InsufficientAmountErrCode, "[fisher] insufficient amount")
	DBFailedErr           = New(DBFailedErrCode, "[fisher] db failed")
)

type TransferErr struct {
	Code ErrCode
	Msg  string
}

func (e *TransferErr) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

func New(code ErrCode, msg string) error {
	return &TransferErr{
		Code: code,
		Msg:  msg,
	}
}

func NewWithErr(code ErrCode, err error) error {
	return &TransferErr{
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
	var e *TransferErr
	if errors.As(err, &e) {
		if e.Code == target.(*TransferErr).Code {
			return true
		}
	}
	return false
}
