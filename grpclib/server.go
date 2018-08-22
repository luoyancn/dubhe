package grpclib

import (
	"context"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"golang.org/x/net/netutil"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/tap"

	"github.com/luoyancn/dubhe/grpclib/config"
	"github.com/luoyancn/dubhe/logging"
)

var once sync.Once
var _grpc *grpc.Server

type registerfunc func(string, bool, string) error
type unregisterfunc func()

var unregist unregisterfunc

type serviceDescKV struct {
	service_server interface{}
	service_desc   grpc.ServiceDesc
}

func NewServiceDescKV(service_server interface{},
	service_desc grpc.ServiceDesc) *serviceDescKV {
	return &serviceDescKV{
		service_server: service_server,
		service_desc:   service_desc,
	}
}

func StartServer(port int, register registerfunc, unregister unregisterfunc,
	service_name string, entities ...*serviceDescKV) {
	once.Do(func() {
		unregist = unregister
		logging.LOG.Infof("Start grpc server and listen on %d\n", port)
		logging.LOG.Infof("%v\n", unregister)
		listener, err := net.Listen(
			"tcp", fmt.Sprintf("%s:%d", "0.0.0.0", port))
		if nil != err {
			logging.LOG.Panicf(
				"Cannot start grpc server on port %d:%v\n",
				port, err)
		}

		var opts []grpc.ServerOption
		if config.GRPC_USE_TLS {
			creds, err := credentials.NewServerTLSFromFile(
				config.GRPC_CA_FILE, config.GRPC_KEY_FILE)
			if nil != err {
				logging.LOG.Fatalf(
					"Cannot init creds for server:%v\n", err)
			}
			opts = append(opts, grpc.Creds(creds))
		}

		if config.GRPC_DEBUG {
			opts = append(opts, withServerDebugInterceptor())
		}

		opts = append(opts, grpc.MaxConcurrentStreams(
			uint32(config.GRPC_CONCURRENCY)))

		if config.GRPC_REQ_MAX_FREQUENCY < math.MaxFloat64 {
			opts = append(opts, grpc.InTapHandle(newTap().handler))
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
		registed := 0
		if config.GRPC_LB_MODE && nil != register {
			logging.LOG.Infof(
				"Running grpc cluster with load balance mode\n")
			logging.LOG.Infof("And the registed address are :%v\n",
				config.GRPC_REGISTERED_ADDRESS)
			logging.LOG.Warningf(
				"Because of lb, delete 127.0.0.1 and 0.0.0.0 \n")
			for _, addr := range config.GRPC_REGISTERED_ADDRESS {
				if "127.0.0.1" == addr || "0.0.0.0" == addr {
					continue
				}
				register(fmt.Sprintf("%s:%d", addr, port),
					config.GRPC_USE_DEPRECATED_LB, service_name)
				registed++
			}
			if 0 == registed {
				logging.LOG.Fatalf(
					"Please using available address \n")
			}
		}

		for _, entity := range entities {
			_grpc.RegisterService(
				&entity.service_desc, entity.service_server)
		}
		if config.GRPC_CONNECTION_LIMIT > 0 {
			listener = netutil.LimitListener(
				listener, config.GRPC_CONNECTION_LIMIT)
		}
		_grpc.Serve(listener)
	})
}

func withServerDebugInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(debugServerInterceptor)
}

func debugServerInterceptor(ctx context.Context,
	req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	logging.LOG.Debugf("Invoke server method=%s duration=%s error=%v",
		info.FullMethod, time.Since(start), err)
	return resp, err
}

type tapp struct {
	lim *rate.Limiter
}

// 限制访问频率，默认无限制
func newTap() *tapp {
	return &tapp{rate.NewLimiter(rate.Limit(config.GRPC_REQ_MAX_FREQUENCY),
		config.GRPC_REQ_BURST_FREQUENCY)}
}

func (t *tapp) handler(ctx context.Context,
	info *tap.Info) (context.Context, error) {
	if !t.lim.Allow() {
		return nil, status.Errorf(
			codes.ResourceExhausted, "Service is over rate limit")
	}
	return ctx, nil
}

func StopServer() {
	if nil != _grpc {
		logging.LOG.Infof("Stop the grpc server...\n")
		_grpc.Stop()
	}
	logging.LOG.Debugf("%v:%v...\n", unregist, config.GRPC_LB_MODE)
	if nil != unregist && config.GRPC_LB_MODE {
		logging.LOG.Infof("Call back the unregister func...\n")
		unregist()
	}
}
