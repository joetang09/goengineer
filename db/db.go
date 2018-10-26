package db

import (
	"errors"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
)

var (
	rander = rand.New(rand.NewSource(time.Now().Unix()))

	dbHolder = map[string]*Wrapper{}

	errDBNotFound = errors.New("DB Not Found")

	errConfig = errors.New("Config Error")
)

type Config map[string]struct {
	Driver          string
	Source          string
	ConnMaxLifeTime int // in second
	MaxIdleConns    int
	MaxOpenConns    int
	Slave           []struct{ Source string }
}

type Cpnt struct{}

func (Cpnt) Init(options ...interface{}) (err error) {

	if len(options) == 0 {
		return
	}

	c, ok := options[0].(*Config)
	if !ok {
		err = errConfig
		return
	}

	for name, config := range *c {
		w := new(Wrapper)
		w.dsn, err = gorm.Open(config.Driver, config.Source)
		if err != nil {
			return
		}
		w.dsn.DB().SetConnMaxLifetime(time.Duration(config.ConnMaxLifeTime) * time.Second)
		w.dsn.DB().SetMaxIdleConns(config.MaxIdleConns)
		w.dsn.DB().SetMaxOpenConns(config.MaxOpenConns)
		for _, s := range config.Slave {
			var slave *gorm.DB

			slave, err = gorm.Open(config.Driver, s.Source)
			if err != nil {
				return
			}
			slave.DB().SetConnMaxLifetime(time.Duration(config.ConnMaxLifeTime) * time.Second)
			slave.DB().SetMaxIdleConns(config.MaxIdleConns)
			slave.DB().SetMaxOpenConns(config.MaxOpenConns)
			w.slave = append(w.slave, slave)
		}

		dbHolder[name] = w
	}

	registerCallback()
	return nil
}

func (Cpnt) CfgKey() string {
	return "db"
}

func (Cpnt) CfgType() interface{} {
	return Config{}
}

func (Cpnt) CfgUpdate(interface{}) {

}

type Wrapper struct {
	dsn   *gorm.DB
	slave []*gorm.DB
}

func (db *Wrapper) Write() *gorm.DB {
	return db.dsn
}

func (db *Wrapper) Read() *gorm.DB {
	if len(db.slave) == 0 {
		return db.Write()
	}
	return db.slave[rander.Intn(len(db.slave))]
}

func Read(name string) (*gorm.DB, error) {

	if w, e := get(name); e == nil {
		return w.Read(), nil
	} else {
		return nil, e
	}
}

func Write(name string) (*gorm.DB, error) {
	if w, e := get(name); e == nil {
		return w.Write(), nil
	} else {
		return nil, e
	}
}

func MustRead(name string) *gorm.DB {
	return mustGet(name).Read()
}

func MustWrite(name string) *gorm.DB {
	return mustGet(name).Write()
}

func mustGet(name string) *Wrapper {
	c, ok := dbHolder[name]
	if !ok {
		panic(errDBNotFound)
	}

	return c
}

func get(name string) (*Wrapper, error) {

	c, ok := dbHolder[name]
	if !ok {
		return nil, errDBNotFound
	}

	return c, nil
}

func registerCallback() {
	gorm.DefaultCallback.Create().After("gorm:update_time_stamp").Register("my:update_time_stamp", func(scope *gorm.Scope) {
		if !scope.HasError() {
			now := time.Now().Unix()
			if ct, ok := scope.FieldByName("CreateTime"); ok {
				ct.Set(now)
			}
			if ct, ok := scope.FieldByName("UpdateTime"); ok {
				ct.Set(now)
			}
		}
	})
	gorm.DefaultCallback.Update().After("gorm:update_time_stamp").Register("my:update_time_stamp", func(scope *gorm.Scope) {
		if _, ok := scope.Get("gorm:update_column"); !ok {
			scope.SetColumn("UpdateTime", time.Now().Unix())
		}
	})
}

func NotDeletedScope(db *gorm.DB) *gorm.DB {
	return db.Where("del_flag = ?", 0)
}

func NotDeletedScopeWithPrefix(prefix ...string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, v := range prefix {
			db = db.Where(v+".del_flag = ?", 0)
		}
		return db
	}
}

func IsError(e error) error {
	if e != nil && e != gorm.ErrRecordNotFound {
		return e
	}
	return nil
}
