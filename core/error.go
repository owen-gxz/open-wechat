package core

import (
	"fmt"
)

const (
	errorErrCodeIndex = 0
	errorErrMsgIndex  = 1
)

type Error struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (err *Error) Error() string {
	return fmt.Sprintf("errcode: %d, errmsg: %s", err.ErrCode, err.ErrMsg)
}

type H map[string]interface{}
