package grpclib

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/resolver"

	"github.com/luoyancn/dubhe/common"
	"github.com/luoyancn/dubhe/grpclib/config"
	"github.com/luoyancn/dubhe/logging"
	"github.com/luoyancn/dubhe/registry/etcdv3"
)

type grpcPool struct {
	conn chan *grpc.ClientConn
	addr string
}

var gonce sync.Once
var pool *grpcPool
var servie_name string

type resolvfunc func() naming.Resolver

var fn resolvfunc

func InitGrpcClientPool(endpoint string, srv string, res resolvfunc) {
	gonce.Do(func() {
		fn = res
		servie_name = srv
		pool = new(grpcPool)
		pool.addr = endpoint
		pool.conn = make(chan *grpc.ClientConn, 1024)
		conn := pool.dialNew()
		pool.conn <- conn
	})
}

func withClientDebugInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(clientDebugInterceptor)
}

func clientDebugInterceptor(ctx context.Context,
	method string, req interface{}, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	logging.LOG.Infof("Invoke remote method=%s duration=%s error=%v",
		method, time.Since(start), err)
	return err
}

func (this *grpcPool) dialNew() *grpc.ClientConn {
	var err error
	var conn *grpc.ClientConn
	opts := []grpc.DialOption{}
	if config.GRPC_USE_TLS {
		creds, err := credentials.NewClientTLSFromFile(config.GRPC_CA_FILE, "")
		if nil != err {
			logging.LOG.Fatalf("Cannot connect to grpc server :%v\n", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if config.GRPC_DEBUG {
		opts = append(opts, withClientDebugInterceptor())
	}

	ctx, cancle := context.WithTimeout(
		context.Background(), config.GRPC_TIMEOUT)
	defer cancle()
	if config.GRPC_LB_MODE {
		logging.LOG.Infof("Using lb mode to visit grpc server...\n")
		if config.GRPC_USE_DEPRECATED_LB {
			logging.LOG.Debugf("grpc.RoundRobin will deprecated, Please use grpc.WithBalancerName instead...\n")
			balancer := grpc.RoundRobin(fn())
			opts = append(opts, grpc.WithBlock(), grpc.WithBalancer(balancer))
			conn, err = grpc.DialContext(ctx, "", opts...)
		} else {
			builder := etcdv3.NewBuilder()
			resolver.Register(builder)
			opts = append(opts, grpc.WithBlock(), grpc.WithBalancerName("round_robin"))
			conn, err = grpc.DialContext(
				ctx, builder.Scheme()+"://"+common.AUTHORITY+"/"+servie_name, opts...)
		}
	} else {
		logging.LOG.Infof("Visit grpc server directly...\n")
		//conn, err = grpc.Dial(this.addr, opts...)
		conn, err = grpc.DialContext(ctx, this.addr, opts...)
	}

	if nil != err {
		logging.LOG.Fatalf("Cannot connect to grpc server :%v\n", err)
	}
	return conn
}

func Get() *grpc.ClientConn {
	return pool.get()
}
func (this *grpcPool) get() *grpc.ClientConn {
	select {
	case conn := <-this.conn:
		return conn
	default:
		return this.dialNew()
	}
}

func Put(conn *grpc.ClientConn) error {
	return pool.put(conn)
}

func (this *grpcPool) put(conn *grpc.ClientConn) error {
	select {
	case this.conn <- conn:
		return nil
	default:
		return conn.Close()
	}
}
