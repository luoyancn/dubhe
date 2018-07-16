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
	viper.Set("etcd.endpoints", []string{"http://localhost:2379"})
	viper.Set("etcd.connection_timeout", 5)
	viper.Set("etcd.ttl", 10)
	viper.Set("etcd.service_name", "etcd")
	viper.Set("etcd.register_dir", "/grpclib")
}

func (this *etcdConf) OverWriteConfig() {
	over_once.Do(func() {
		ETCD_ENDPOINTS = viper.GetStringSlice("etcd.endpoints")
		ETCD_CONNECTION_TIMEOUT = viper.GetDuration(
			"etcd.connection_timeout") * time.Second
		ETCD_TTL = viper.GetDuration("etcd.ttl") * time.Second
		ETCD_SERVICE_NAME = viper.GetString("etcd.service_name")
		ETCD_REGISTER_DIR = viper.GetString("etcd.register_dir")
	})
}
