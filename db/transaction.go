package db

import (
	"fmt"
	"runtime/debug"

	"github.com/jinzhu/gorm"
)

const (
	intErrCode = 1
	dbErrCode  = 2

	transErrF = "TransError[%d]: %s"
)

type TransError struct {
	Code int
	Msg  string
}

func (t TransError) Error() string {
	return fmt.Sprintf(transErrF, t.Code, t.Msg)
}

type Task func(db *gorm.DB) error

func ErrHandler(db *gorm.DB, task Task) (err error) {
	defer func() {
		if e := recover(); e != nil {
			msg := fmt.Sprintf("panic: %s\ncalltrace : %s", fmt.Sprint(e), string(debug.Stack()))
			err = &TransError{intErrCode, msg}
		}
	}()
	return task(db)
}

func ExecTrans(db *gorm.DB, trans ...Task) error {
	execDb := db.Begin()
	if execDb.Error != nil {
		fmt.Printf("DB begin transaction failed: %s", execDb.Error.Error())
		return &TransError{dbErrCode, execDb.Error.Error()}
	}
	for _, task := range trans {
		if err := ErrHandler(execDb, task); err != nil {
			if err := execDb.Rollback().Error; err != nil {
				fmt.Printf("roll_back : %s", execDb.Error.Error())
			}
			return err
		}
	}

	if err := execDb.Commit().Error; err != nil {
		execDb.Rollback()
		return &TransError{dbErrCode, err.Error()}
	}
	return nil
}
