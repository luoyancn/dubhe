package etcdv3

import (
	"github.com/luoyancn/dubhe/logging"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"
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
		resp, err := this.client.Get(
			context.Background(), this.key, etcd.WithPrefix())
		if err == nil {
			addrs := extractAddrs(resp)
			if len(addrs) > 0 {
				for _, addr := range addrs {
					updates = append(updates, &naming.Update{
						Op: naming.Add, Addr: addr})
				}
				this.updates = updates
				return updates, nil
			}
		}
	}

	rch := this.client.Watch(
		context.Background(), this.key, etcd.WithPrefix())
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
