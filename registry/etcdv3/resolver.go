package etcdv3

import (
	"errors"
	"fmt"

	"github.com/luoyancn/dubhe/logging"

	etcd "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

type etcdResolver struct {
	config      etcd.Config
	registryDir string
	serviceName string
}

func NewResolver(registryDir string, serviceName string,
	cfg etcd.Config) naming.Resolver {
	if cfg.DialTimeout > 0 {
		logging.LOG.Warningf("With DialTimeout options, we must use blocking\n")
		logging.LOG.Warningf(" call to Ensure the error would be returned\n")
		logging.LOG.Warningf(" Visit the follow links to get more details:\n")
		logging.LOG.Warningf(" https://github.com/coreos/etcd/issues/9829\n")
		logging.LOG.Warningf(" https://github.com/coreos/etcd/issues/9877\n")
		cfg.DialOptions = append(cfg.DialOptions, grpc.WithBlock())
	}
	return &etcdResolver{registryDir: registryDir,
		serviceName: serviceName, config: cfg}
}

func (this *etcdResolver) Resolve(target string) (naming.Watcher, error) {
	if this.serviceName == "" {
		return nil, errors.New("no service name provided")
	}
	client, err := etcd.New(this.config)
	if err != nil {
		logging.LOG.Errorf(
			"Cannot connect to endpoins :%v\n", this.config.Endpoints)
		return nil, err
	}
	key := fmt.Sprintf("%s/%s", this.registryDir, this.serviceName)
	logging.LOG.Debugf("Watching the key named:%s\n", key)
	return newEtcdWatcher(key, client), nil
}
