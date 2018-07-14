package etcdv3

import (
	"strings"

	"github.com/luoyancn/dubhe/logging"
	etcdconf "github.com/luoyancn/dubhe/registry/etcdv3/config"

	etcd "github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

	// 1.no dial timeout means New doesn't block and connection attempt
	// happens in the background.
	// 2.dial timeout to wait up to a fixed amount of time until connection up.
	// Because of those above, if the config of etcd used DialTimeout, we must
	// append grpc.WithBlock into the DialOptions of config. Otherwise,
	// lientv3.New() won't return error when no endpoint is available.
	// https://github.com/coreos/etcd/issues/9829
	// https://github.com/coreos/etcd/issues/9877
	if config.DialTimeout > 0 {
		logging.LOG.Warningf("With DialTimeout options, we must use blocking\n")
		logging.LOG.Warningf(" call to Ensure the error would be returned\n")
		logging.LOG.Warningf(" Visit the follow links to get more details:\n")
		logging.LOG.Warningf(" https://github.com/coreos/etcd/issues/9829\n")
		logging.LOG.Warningf(" https://github.com/coreos/etcd/issues/9877\n")
		config.DialOptions = append(config.DialOptions, grpc.WithBlock())
	}
	client, err := etcd.New(config)
	if err != nil {
		logging.LOG.Fatalf("Cannot connect to endpoins :%v in %d seconds\n",
			config.Endpoints, config.DialTimeout)
	}

	go func() {
		<-deregister
		logging.LOG.Infof("Deleting %s from etcd...\n", key)
		client.Delete(context.Background(), key)
		deregister <- struct{}{}
	}()

	grant_ctx, grant_cancle := context.WithTimeout(
		context.Background(), etcdconf.ETCD_CONNECTION_TIMEOUT)
	defer grant_cancle()
	resp, err := client.Grant(grant_ctx, int64(etcdconf.ETCD_TTL))
	if err != nil {
		logging.LOG.Errorf("Failed connect to etcd:%v\n", err)
		return err
	}

	put_ctx, put_cancle := context.WithTimeout(
		context.Background(), etcdconf.ETCD_CONNECTION_TIMEOUT)
	defer put_cancle()
	if _, err = client.Put(put_ctx, key,
		opt.NData, etcd.WithLease(resp.ID)); err != nil {
		logging.LOG.Errorf(
			"Failed Put the key %s value %sinto etcd:%v\n", key, opt.NData, err)
		return err
	}

	keep_ctx, keep_cancle := context.WithTimeout(
		context.Background(), etcdconf.ETCD_CONNECTION_TIMEOUT)
	defer keep_cancle()
	if _, err := client.KeepAlive(keep_ctx, resp.ID); err != nil {
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
