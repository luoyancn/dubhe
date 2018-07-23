package config

import (
	"math"
	"sync"
	"time"

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
	viper.SetDefault("grpc.lb_mode", false)
	viper.SetDefault("grpc.registered_address", []string{})
	viper.SetDefault("grpc.debug", false)
	viper.SetDefault("grpc.concurrency", 1024)
	viper.SetDefault("grpc.req_max_frequency", math.MaxFloat64)
	viper.SetDefault("grpc.req_burst_frequency", 10)
	viper.SetDefault("grpc.connection_limit", 10240)
	viper.SetDefault("grpc.timeout", 10)
}

func (this *grpConf) OverWriteConfig() {
	over_once.Do(func() {
		GRPC_PORT = viper.GetInt("grpc.port")
		GRPC_USE_TLS = viper.GetBool("grpc.use_tls")
		GRPC_CA_FILE = viper.GetString("grpc.ca_file")
		GRPC_KEY_FILE = viper.GetString("grpc.key_file")
		GRPC_LB_MODE = viper.GetBool("grpc.lb_mode")
		GRPC_REGISTERED_ADDRESS = viper.GetStringSlice(
			"grpc.registered_address")
		GRPC_DEBUG = viper.GetBool("grpc.debug")
		GRPC_CONCURRENCY = viper.GetInt("grpc.concurrency")
		GRPC_REQ_MAX_FREQUENCY = viper.GetFloat64("grpc.req_max_frequency")
		GRPC_REQ_BURST_FREQUENCY = viper.GetInt("grpc.req_burst_frequency")
		GRPC_CONNECTION_LIMIT = viper.GetInt("grpc.connection_limit")
		GRPC_TIMEOUT = viper.GetDuration("grpc.timeout") * time.Second
	})
}
