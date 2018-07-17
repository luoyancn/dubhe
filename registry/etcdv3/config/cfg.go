package config

import (
	"sync"
	"time"

	"github.com/spf13/viper"
)

var init_once sync.Once
var over_once sync.Once

type etcdConf struct {
}

func NewEtcdConf() *etcdConf {
	return &etcdConf{}
}

func init() {
	init_once.Do(func() {
		set_etcd()
	})
}

func set_etcd() {
	viper.SetDefault("etcd.endpoints", []string{"http://localhost:2379"})
	viper.SetDefault("etcd.connection_timeout", 5)
	viper.SetDefault("etcd.ttl", 10)
	viper.SetDefault("etcd.service_name", "etcd")
	viper.SetDefault("etcd.register_dir", "/grpclib")
	viper.SetDefault("etcd.use_tls", false)
	viper.SetDefault("etcd.ca_cert", "")
	viper.SetDefault("etcd.ca_file", "")
	viper.SetDefault("etcd.key_file", "")
}

func (this *etcdConf) OverWriteConfig() {
	over_once.Do(func() {
		ETCD_ENDPOINTS = viper.GetStringSlice("etcd.endpoints")
		ETCD_CONNECTION_TIMEOUT = viper.GetDuration(
			"etcd.connection_timeout") * time.Second
		ETCD_TTL = viper.GetDuration("etcd.ttl") * time.Second
		ETCD_SERVICE_NAME = viper.GetString("etcd.service_name")
		ETCD_REGISTER_DIR = viper.GetString("etcd.register_dir")
		ETCD_USE_TLS = viper.GetBool("etcd.use_tls")
		ETCD_CA_CERT = viper.GetString("etcd.ca_cert")
		ETCD_CA_FILE = viper.GetString("etcd.ca_file")
		ETCD_KEY_FILE = viper.GetString("etcd.key_file")
	})
}
