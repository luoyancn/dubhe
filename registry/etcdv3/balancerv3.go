package etcdv3

import (
	"context"
	"strings"

	etcd "github.com/coreos/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/resolver"

	"github.com/luoyancn/dubhe/logging"
)

const schema = "zhangzz"

var cli *etcd.Client

type etcdbuilder struct {
	cc resolver.ClientConn
}

func NewResolver(service_name string) (naming.Resolver, resolver.Builder) {

	return &etcdResolver{service_name: service_name}, &etcdbuilder{}
}

func (this *etcdbuilder) Build(target resolver.Target, cc resolver.ClientConn,
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
	this.cc = cc
	key := "/" + target.Scheme + "/" + target.Endpoint + "/"
	logging.LOG.Infof("Building the key with %s\n", key)
	go this.watch(key)
	return this, nil
}

func (this *etcdbuilder) Scheme() string {
	return schema
}

func (this *etcdbuilder) ResolveNow(rn resolver.ResolveNowOption) {
}

func (this *etcdbuilder) Close() {
}

func (this *etcdbuilder) watch(keyPrefix string) {
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

	this.cc.NewAddress(addrList)

	rch := cli.Watch(context.Background(), keyPrefix, etcd.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			addr := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix)
			switch ev.Type {
			case mvccpb.PUT:
				if !exist(addrList, addr) {
					addrList = append(addrList,
						resolver.Address{Addr: addr})
					this.cc.NewAddress(addrList)
				}
			case mvccpb.DELETE:
				if s, ok := remove(addrList, addr); ok {
					addrList = s
					this.cc.NewAddress(addrList)
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
