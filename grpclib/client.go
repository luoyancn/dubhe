package grpclib

import (
	"sync"

	"github.com/luoyancn/dubhe/grpclib/config"
	"github.com/luoyancn/dubhe/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type grpcPool struct {
	conn chan *grpc.ClientConn
	addr string
}

var gonce sync.Once
var pool *grpcPool

func InitGrpcClientPool(endpoint string) {
	gonce.Do(func() {
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
	if config.GRPC_USE_TLS {
		creds, err := credentials.NewClientTLSFromFile(config.GRPC_CA_FILE, "")
		if nil != err {
			logging.LOG.Fatalf("Cannot connect to grpc server :%v\n", err)
		}
		conn, err = grpc.Dial(this.addr, grpc.WithTransportCredentials(creds))
	} else {
		conn, err = grpc.Dial(this.addr, grpc.WithInsecure())
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
