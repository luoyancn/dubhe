package etcdv3

import (
	"strings"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/luoyancn/dubhe/logging"
	etcdconf "github.com/luoyancn/dubhe/registry/etcdv3/config"
)

var deregister = make(chan struct{})

func generate_etcd_config() etcd.Config {
	config := etcd.Config{
		Endpoints:   etcdconf.ETCD_ENDPOINTS,
		DialTimeout: etcdconf.ETCD_CONNECTION_TIMEOUT,
	}

	// 1.no dial timeout means New doesn't block and connection attempt
	// happens in the background.
	// 2.dial timeout to wait up to a fixed amount of time until connection up.
	// Because of those above, if the config of etcd used DialTimeout, we must
	// append grpc.WithBlock into the DialOptions of config. Otherwise,
	// lientv3.New() won't return error when no endpoint is available.
	// https://github.com/coreos/etcd/issues/9829
	// https://github.com/coreos/etcd/issues/9877
	if etcdconf.ETCD_CONNECTION_TIMEOUT > 0 {
		logging.LOG.Warningf(
			"With DialTimeout options, we must use blocking\n")
		logging.LOG.Warningf(
			"call to Ensure the error would be returned\n")
		logging.LOG.Warningf(
			"Visit the follow links to get more details:\n")
		logging.LOG.Warningf(
			"https://github.com/coreos/etcd/issues/9829\n")
		logging.LOG.Warningf(
			"https://github.com/coreos/etcd/issues/9877\n")
		config.DialOptions = append(
			config.DialOptions, grpc.WithBlock())
	}

	if etcdconf.ETCD_USE_TLS {
		logging.LOG.Warningf("Conneciting etcd with tls...\n")
		tls := transport.TLSInfo{
			TrustedCAFile: etcdconf.ETCD_CA_CERT,
			CertFile:      etcdconf.ETCD_CA_FILE,
			KeyFile:       etcdconf.ETCD_KEY_FILE,
		}
		tlsConfig, err := tls.ClientConfig()
		if nil != err {
			logging.LOG.Fatalf("Cannot init tls for etcd client:%v\n", err)
		}
		config.TLS = tlsConfig
	}
	return config
}

func Register(ndata string, deprecated bool, service_name string) error {
	config := generate_etcd_config()
	nodeid, _ := uuid.NewV4()
	key := "/" + schema + "/" + service_name + "/" + ndata
	if deprecated {
		logging.LOG.Warningf("Register service with deprecated lb mode\n")
		key = strings.Join(
			[]string{etcdconf.ETCD_REGISTER_DIR,
				service_name, nodeid.String()}, "/")
	}
	logging.LOG.Infof("Register service with key :%s\n", key)

	var client *etcd.Client
	go func() {
		<-deregister
		if nil != client {
			logging.LOG.Infof("Deleting %s from etcd...\n", key)
			client.Delete(context.Background(), key)
		}
		deregister <- struct{}{}
	}()

	client, err := etcd.New(config)
	if err != nil {
		logging.LOG.Fatalf("Cannot connect to endpoins %v in %d seconds:%v\n",
			config.Endpoints, config.DialTimeout/time.Second, err)
	}

	grant_ctx, grant_cancle := context.WithTimeout(
		context.Background(), etcdconf.ETCD_CONNECTION_TIMEOUT)
	defer grant_cancle()
	// The second param of Grant is second, not Duration
	// we must convert from duration to second
	// And the max ttl is 9,000,000,000
	// See more detail:
	// https://github.com/coreos/etcd/blob/master/clientv3/options.go#L65
	resp, err := client.Grant(grant_ctx, int64(etcdconf.ETCD_TTL/time.Second))
	if err != nil {
		logging.LOG.Errorf("Failed connect to etcd:%v\n", err)
		return err
	}

	put_ctx, put_cancle := context.WithTimeout(
		context.Background(), etcdconf.ETCD_CONNECTION_TIMEOUT)
	defer put_cancle()
	logging.LOG.Debugf("Put key %s value %s into etcd...\n", key, ndata)
	if _, err = client.Put(put_ctx, key,
		ndata, etcd.WithLease(resp.ID)); err != nil {
		logging.LOG.Errorf(
			"Failed Put the key %s value %sinto etcd:%v\n",
			key, ndata, err)
		return err
	}

	// Because we must ensure the key always alive in etcd,
	// we must use context without timeout. Otherwise, the
	// grpclient cannot receive any response from grpcserver after
	// timeout.
	keep, err := client.KeepAlive(context.TODO(), resp.ID)
	if nil != err {
		logging.LOG.Errorf(
			"Failed refresh the key %s exsit in etcd:%v\n",
			key, err)
		return err
	}

	// Eat all keep alive messages
	// Otherwise, lease keepalive response queue will full,
	// and etcd would dropping response send
	// For more detail, visit the follow links:
	// https://github.com/coreos/etcd/pull/9952
	// https://github.com/coreos/etcd/issues/8168
	// https://github.com/coreos/etcd/blob/master/clientv3/concurrency/session.go#L63
	// https://www.cnxct.com/etcd-lease-keepalive-debug-note/

	go func() {
		for range keep {
		}
	}()
	return nil
}

func UnRegister() {
	deregister <- struct{}{}
	<-deregister
}
