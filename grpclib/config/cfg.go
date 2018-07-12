package config

import (
	"sync"

	"github.com/spf13/viper"
)

var init_once sync.Once
var over_once sync.Once

type grpConf struct {
}

func NewGrpConf() *grpConf {
	return &grpConf{}
}

func init() {
	init_once.Do(func() {
		set_grpc()
	})
}

func set_grpc() {
	viper.Set("grpc.port", 8080)
}

func (this *grpConf) OverWriteConfig() {
	over_once.Do(func() {
		GRPC_PORT = viper.GetInt("grpc.port")
	})
}
