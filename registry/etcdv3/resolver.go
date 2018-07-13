package etcdv3

import (
	"errors"
	"fmt"

	"github.com/luoyancn/dubhe/logging"

	etcd "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/naming"
)

type etcdResolver struct {
	config      etcd.Config
	registryDir string
	serviceName string
}

func NewResolver(registryDir string, serviceName string,
	cfg etcd.Config) naming.Resolver {
	return &etcdResolver{registryDir: registryDir,
		serviceName: serviceName, config: cfg}
}

func (this *etcdResolver) Resolve(target string) (naming.Watcher, error) {
	if this.serviceName == "" {
		return nil, errors.New("no service name provided")
	}
	client, err := etcd.New(this.config)
	if err != nil {
		logging.LOG.Errorf("Cannot init etcd cluster:%v\n", err)
		return nil, err
	}

	key := fmt.Sprintf("%s/%s", this.registryDir, this.serviceName)
	logging.LOG.Debugf("Watching the key named:%s\n", key)
	return newEtcdWatcher(key, client), nil
}
