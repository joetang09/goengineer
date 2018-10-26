package engineer

import (
	"reflect"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultConfigFile = "config.toml"
)

var (
	cf        = new(string)
	configIns = config{vip: viper.New()}
	configPkg = reflect.TypeOf(configIns).PkgPath() + ".ConfigCpnt"
)

type config struct {
	vip      *viper.Viper
	filePath string
}

type ConfigCpnt struct{}

func (ConfigCpnt) Init(options ...interface{}) error {

	if *cf == defaultConfigFile && len(options) > 0 {
		c, ok := options[0].(string)
		if ok {
			*cf = c
		}
	}

	return configIns.SetConfigFile(*cf)

}

func (ConfigCpnt) CfgKey() string {
	return ""

}

func (ConfigCpnt) CfgType() interface{} {
	return struct{}{}
}

func (ConfigCpnt) CfgUpdate(interface{}) {

}

func (c config) SetConfigFile(f string) error {
	c.vip.SetConfigFile(f)
	c.vip.AutomaticEnv()

	return c.vip.ReadInConfig()
}

func (c config) Get(key string) interface{} {
	return c.vip.Get(key)
}

func (c config) GetBool(key string) bool {
	return c.vip.GetBool(key)
}

func (c config) GetDuration(key string) time.Duration {
	return c.vip.GetDuration(key)
}

func (c config) GetFloat64(key string) float64 {
	return c.vip.GetFloat64(key)
}

func (c config) GetInt(key string) int {
	return c.vip.GetInt(key)
}

func (c config) GetInt64(key string) int64 {
	return c.vip.GetInt64(key)
}

func (c config) GetSizeInBytes(key string) uint {
	return c.vip.GetSizeInBytes(key)
}

func (c config) GetString(key string) string {
	return c.vip.GetString(key)
}

func (c config) GetStringMap(key string) map[string]interface{} {
	return c.vip.GetStringMap(key)
}

func (c config) GetStringMapString(key string) map[string]string {
	return c.vip.GetStringMapString(key)
}

func (c config) GetStringMapStringSlice(key string) map[string][]string {
	return c.vip.GetStringMapStringSlice(key)
}

func (c config) GetStringSlice(key string) []string {
	return c.vip.GetStringSlice(key)
}

func (c config) GetTime(key string) time.Time {
	return c.vip.GetTime(key)
}

func (c config) IsSet(key string) bool {
	return c.vip.IsSet(key)
}
