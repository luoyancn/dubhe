package etcdv3

import (
	"errors"
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/naming"

	"github.com/luoyancn/dubhe/logging"
	etcdconf "github.com/luoyancn/dubhe/registry/etcdv3/config"
)

type etcdResolver struct {
	service_name string
}

func (this *etcdResolver) Resolve(target string) (naming.Watcher, error) {
	if this.service_name == "" {
		return nil, errors.New("no service name provided")
	}

	config := generate_etcd_config()
	client, err := etcd.New(config)
	if err != nil {
		logging.LOG.Errorf(
			"Cannot connect to endpoins :%v\n", config.Endpoints)
		return nil, err
	}
	key := fmt.Sprintf(
		"%s/%s", etcdconf.ETCD_REGISTER_DIR, this.service_name)
	logging.LOG.Debugf("Watching the key named:%s\n", key)
	return newEtcdWatcher(key, client), nil
}
