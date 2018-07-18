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
	viper.SetDefault("grpc.port", 8080)
	viper.SetDefault("grpc.use_tls", false)
	viper.SetDefault("grpc.ca_file", "")
	viper.SetDefault("grpc.key_file", "")
}

func (this *grpConf) OverWriteConfig() {
	over_once.Do(func() {
		GRPC_PORT = viper.GetInt("grpc.port")
		GRPC_USE_TLS = viper.GetBool("grpc.use_tls")
		GRPC_CA_FILE = viper.GetString("grpc.ca_file")
		GRPC_KEY_FILE = viper.GetString("grpc.key_file")
	})
}
