package grpclib

import (
	"sync"

	"google.golang.org/grpc"
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
	conn, err = grpc.Dial(this.addr, grpc.WithInsecure())
	if nil != err {
		return nil
	}
	return conn
}

func Get() *grpc.ClientConn {
	return pool.Get()
}
func (this *grpcPool) Get() *grpc.ClientConn {
	select {
	case conn := <-this.conn:
		return conn
	default:
		return this.dialNew()
	}
}

func Put(conn *grpc.ClientConn) error {
	return pool.Put(conn)
}

func (this *grpcPool) Put(conn *grpc.ClientConn) error {
	select {
	case this.conn <- conn:
		return nil
	default:
		return conn.Close()
	}
}
