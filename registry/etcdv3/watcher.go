package etcdv3

import (
	etcd "github.com/coreos/etcd/clientv3"
	_ "github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"

	"github.com/luoyancn/dubhe/logging"
	etcdconf "github.com/luoyancn/dubhe/registry/etcdv3/config"
)

type etcdWatcher struct {
	key     string
	client  *etcd.Client
	updates []*naming.Update
}

func (this *etcdWatcher) Close() {
}

func newEtcdWatcher(key string, cli *etcd.Client) naming.Watcher {
	this := &etcdWatcher{
		key:     key,
		client:  cli,
		updates: make([]*naming.Update, 0),
	}
	return this
}

func (this *etcdWatcher) Next() ([]*naming.Update, error) {
	updates := make([]*naming.Update, 0)

	if len(this.updates) == 0 {
		get_ctx, get_cancle := context.WithTimeout(
			context.Background(), etcdconf.ETCD_CONNECTION_TIMEOUT)
		defer get_cancle()
		resp, err := this.client.Get(
			get_ctx, this.key, etcd.WithPrefix())
		if err == nil {
			addrs := extractAddrs(resp)
			if len(addrs) > 0 {
				for _, addr := range addrs {
					updates = append(updates,
						&naming.Update{Op: naming.Add,
							Addr: addr})
				}
				this.updates = updates
				return updates, nil
			}
		} else {
			logging.LOG.Warningf(
				"Cannot get entity from etcd with key :%s\n",
				this.key)
		}
	}

	watch_ctx, watch_cancle := context.WithTimeout(
		context.Background(), etcdconf.ETCD_CONNECTION_TIMEOUT)
	defer watch_cancle()
	rch := this.client.Watch(watch_ctx, this.key, etcd.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			addr := ev.Kv.String()
			switch ev.Type {
			case mvccpb.PUT:
				updates = append(updates, &naming.Update{
					Op: naming.Add, Addr: addr})
			case mvccpb.DELETE:
				updates = append(updates, &naming.Update{
					Op: naming.Delete, Addr: addr})
			}
		}
	}
	return updates, nil
}

func extractAddrs(resp *etcd.GetResponse) []string {
	addrs := []string{}

	if resp == nil || resp.Kvs == nil {
		return addrs
	}

	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			addrs = append(addrs, string(v))
		}
	}

	logging.LOG.Debugf("The service endpoint is %v\n", addrs)
	return addrs
}
