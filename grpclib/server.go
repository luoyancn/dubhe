package grpclib

import (
	"fmt"
	"net"
	"sync"

	"github.com/luoyancn/dubhe/grpclib/config"
	"github.com/luoyancn/dubhe/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var once sync.Once
var _grpc *grpc.Server

func StartServer(port int, ss interface{}, desc ...grpc.ServiceDesc) {
	once.Do(func() {
		logging.LOG.Infof("Start grpc server and listen on %d\n", port)
		listener, err := net.Listen(
			"tcp", fmt.Sprintf("%s:%d", "0.0.0.0", port))
		if nil != err {
			logging.LOG.Panicf(
				"Cannot start grpc server on port %d:%v\n", port, err)
		}

		var opts []grpc.ServerOption
		if config.GRPC_USE_TLS {
			creds, err := credentials.NewServerTLSFromFile(
				config.GRPC_CA_FILE, config.GRPC_KEY_FILE)
			if nil != err {
				logging.LOG.Fatalf(
					"Cannot init creds for grpc server:%v\n", err)
			}
			opts = append(opts, grpc.Creds(creds))
		}
		_grpc = grpc.NewServer(opts...)
		// In general, we registe service into grpc like follows:
		// messages.RegisterReQRePServer(_grpc,
		// &messages.Service{HostName: host_name, ListenPort: port})

		// While, for split grpc server and service, we could use follows
		// in codes. Be care for these above, if we use like follows, we
		// must modify the pb.go files manually, only export _xxx_serviceDesc
		// to xxx_serviceDesc. Remind, the second params is the grpc service
		// entity.
		// var xxx_serviceDesc = _xxx_serviceDesc
		for _, d := range desc {
			_grpc.RegisterService(&d, ss)
		}
		_grpc.Serve(listener)
	})
}

func StopServer() {
	logging.LOG.Infof("Stop the grpc server...\n")
	if nil != _grpc {
		_grpc.Stop()
	}
}
