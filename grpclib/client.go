package grpclib

import (
	"sync"

	"github.com/luoyancn/dubhe/grpclib/config"
	"github.com/luoyancn/dubhe/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/naming"
)

type grpcPool struct {
	conn chan *grpc.ClientConn
	addr string
}

var gonce sync.Once
var pool *grpcPool

type resolvfunc func() naming.Resolver

var fn resolvfunc

func InitGrpcClientPool(endpoint string, res resolvfunc) {
	gonce.Do(func() {
		fn = res
		pool = new(grpcPool)
		pool.addr = endpoint
		pool.conn = make(chan *grpc.ClientConn, 1024)
		conn := pool.dialNew()
		pool.conn <- conn
	})
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

	if config.GRPC_LB_MODE {
		logging.LOG.Infof("Using lb mode to visit grpc server...\n")
		balancer := grpc.RoundRobin(fn())
		opts = append(opts, grpc.WithBlock(), grpc.WithBalancer(balancer))
		conn, err = grpc.Dial("", opts...)
	} else {
		logging.LOG.Infof("Visit grpc server directly...\n")
		conn, err = grpc.Dial(this.addr, opts...)
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
