package engineer

import (
	"reflect"
	"sync"

	"github.com/mitchellh/mapstructure"
)

var (
	cpntBox sync.Map
)

const (
	useCannotBeAnPointer = "Use(Component) Cannot be an Pointer"
	useCannotToBeTwice   = "Use(Component) Cannot to be Twice"
)

type Component interface {
	Init(...interface{}) error
	CfgKey() string
	CfgType() interface{}
	CfgUpdate(interface{})
}

func Use(c Component, options ...interface{}) {

	t := reflect.TypeOf(c)
	if t.Kind() == reflect.Ptr {
		panic(useCannotBeAnPointer)
	}

	compKey := t.PkgPath() + "." + t.Name()

	if len(options) == 0 && compKey != configPkg && IsUsed(configPkg) {

		ct := c.CfgType()
		if ct != nil {
			cmt := reflect.TypeOf(c.CfgType())
			cm := reflect.New(cmt).Interface()

			if err := mapstructure.WeakDecode(configIns.Get(c.CfgKey()), cm); err == nil {
				options = append(options, cm)
			} else {
				enginerLogger.Info(compKey, " Parse Config Error : ", err)
			}
		}

	}

	if _, ok := cpntBox.LoadOrStore(compKey, c); ok {
		panic(useCannotToBeTwice)
	}

	if err := c.Init(options...); err != nil {
		enginerLogger.Info("Init ", compKey, err)
		panic(err)
	}

}

func IsUsed(c string) bool {

	_, ok := cpntBox.Load(c)
	return ok

}

func UsedList() []string {

	r := []string{}
	cpntBox.Range(func(k, v interface{}) bool {
		ks := k.(string)
		r = append(r, ks)
		return true
	})
	return r

}
