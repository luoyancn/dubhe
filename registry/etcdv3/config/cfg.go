package config

import (
	"sync"

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
}

func (this *etcdConf) OverWriteConfig() {
	over_once.Do(func() {
		ETCD_ENDPOINTS = viper.GetStringSlice("etcd.endpoints")
	})
}
