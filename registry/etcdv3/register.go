package etcdv3

import (
	"strings"

	"github.com/luoyancn/dubhe/logging"

	etcd "github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

var deregister = make(chan struct{})

type Option struct {
	RegistryDir string
	ServiceName string
	NodeID      string
	NData       string
}

func Register(config etcd.Config, opt Option) error {
	key := strings.Join(
		[]string{opt.RegistryDir, opt.ServiceName, opt.NodeID}, "/")
	logging.LOG.Infof("Register service with key :%s\n", key)
	client, err := etcd.New(config)
	if err != nil {
		panic(err)
	}
	go func() {
		<-deregister
		logging.LOG.Infof("Deleting %s from etcd...\n", key)
		client.Delete(context.Background(), key)
		deregister <- struct{}{}
	}()

	resp, err := client.Grant(context.TODO(), int64(10))
	if err != nil {
		logging.LOG.Errorf("Failed connect to etcd:%v\n", err)
		return err
	}

	if _, err = client.Put(context.TODO(), key,
		opt.NData, etcd.WithLease(resp.ID)); err != nil {
		logging.LOG.Errorf(
			"Failed Put the key %s value %sinto etcd:%v\n", key, opt.NData, err)
		return err
	}

	if _, err := client.KeepAlive(context.TODO(), resp.ID); err != nil {
		logging.LOG.Errorf(
			"Failed refresh the key %s exsit in etcd:%v\n", key, err)
		return err
	}
	return nil
}

func UnRegister() {
	deregister <- struct{}{}
	<-deregister
}
