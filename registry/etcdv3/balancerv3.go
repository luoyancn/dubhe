package etcdv3

import (
	"context"
	"strings"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/resolver"

	"github.com/luoyancn/dubhe/logging"
)

const schema = "zhangzz"

var cli *etcd.Client

type etcdbuilder struct {
	cc resolver.ClientConn
}

func NewResolver() (naming.Resolver, resolver.Builder) {

	return &etcdResolver{}, &etcdbuilder{}
}

func (r *etcdbuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOption) (resolver.Resolver, error) {
	var err error
	if cli == nil {
		config := generate_etcd_config()
		cli, err = etcd.New(config)
		if err != nil {
			logging.LOG.Errorf(
				"Cannot connect to endpoins :%v\n", config.Endpoints)
			return nil, err
		}
	}
	r.cc = cc
	key := "/" + target.Scheme + "/" + target.Endpoint + "/"
	logging.LOG.Infof("Building the key with %s\n", key)
	go r.watch(key)
	return r, nil
}

func (r *etcdbuilder) Scheme() string {
	return schema
}

func (r *etcdbuilder) ResolveNow(rn resolver.ResolveNowOption) {
}

func (r *etcdbuilder) Close() {
}

func (r *etcdbuilder) watch(keyPrefix string) {
	var addrList []resolver.Address

	getResp, err := cli.Get(context.Background(),
		keyPrefix, etcd.WithPrefix())
	if err != nil {
		logging.LOG.Errorf("ERROR:%s\n", err)
	} else {
		for i := range getResp.Kvs {
			addrList = append(addrList, resolver.Address{
				Addr: strings.TrimPrefix(
					string(getResp.Kvs[i].Key), keyPrefix)})
		}
	}

	r.cc.NewAddress(addrList)

	rch := cli.Watch(context.Background(), keyPrefix, etcd.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			addr := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix)
			switch ev.Type {
			case mvccpb.PUT:
				if !exist(addrList, addr) {
					addrList = append(addrList,
						resolver.Address{Addr: addr})
					r.cc.NewAddress(addrList)
				}
			case mvccpb.DELETE:
				if s, ok := remove(addrList, addr); ok {
					addrList = s
					r.cc.NewAddress(addrList)
				}
			}
		}
	}
}

func exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
