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

type reg func(string) error
type unreg func()

var un_fn unreg

func StartServer(port int, ss interface{}, fn reg,
	unfn unreg, desc ...grpc.ServiceDesc) {
	once.Do(func() {
		un_fn = unfn
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

		reg := 0
		if config.GRPC_LB_MODE && nil != fn {
			logging.LOG.Infof("Running grpc cluster with load balance mode\n")
			logging.LOG.Infof("And the registed address are :%v\n",
				config.GRPC_REGISTERED_ADDRESS)
			logging.LOG.Warningf(
				"Because of lb, delete 127.0.0.1 and 0.0.0.0 from address\n")
			for _, addr := range config.GRPC_REGISTERED_ADDRESS {
				if "127.0.0.1" == addr || "0.0.0.0" == addr {
					continue
				}
				fn(fmt.Sprintf("%s:%d", addr, port))
				reg++
			}
			if 0 == reg {
				logging.LOG.Fatalf(
					"Please using available address to regist your service\n")
			}
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
	if nil != un_fn && config.GRPC_LB_MODE {
		un_fn()
	}
}
