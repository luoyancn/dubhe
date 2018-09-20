package config

import (
	"sync"

	"github.com/spf13/viper"
)

var init_once sync.Once
var over_once sync.Once

type defaultConf struct {
}

func NewDefConf() *defaultConf {
	return &defaultConf{}
}

func init() {
	init_once.Do(func() {
		set_default()
	})
}

func set_default() {
	//viper.SetDefault("default.log_path", "/var/log")
	//viper.SetDefault("default.log_prefix", "log")
	viper.SetDefault("default.log_keep", 7)
	viper.SetDefault("default.log_maxsize", 1024*1024*1024*10)
	viper.SetDefault("default.log_debug", false)
}

func (this *defaultConf) OverWriteConfig() {
	over_once.Do(func() {
		LOG_PATH = viper.GetString("default.log_path")
		LOG_PREFIX = viper.GetString("default.log_prefix")
		LOG_KEEP = viper.GetInt("default.log_keep")
		LOG_MAXSIZE = viper.GetInt("default.log_maxsize")
		LOG_DEBUG = viper.GetBool("default.log_debug")
	})
}
