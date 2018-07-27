package grpclib

import (
	"context"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/luoyancn/dubhe/grpclib/config"
	"github.com/luoyancn/dubhe/logging"

	"golang.org/x/net/netutil"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/tap"
)

var once sync.Once
var _grpc *grpc.Server

type reg func(string) error
type unreg func()

var un_fn unreg

type serviceDescKV struct {
	inter interface{}
	desc  grpc.ServiceDesc
}

func NewServiceDescKV(
	inter interface{}, desc grpc.ServiceDesc) *serviceDescKV {
	return &serviceDescKV{
		inter: inter,
		desc:  desc,
	}
}

func StartServer(port int, fn reg, unfn unreg, entities ...*serviceDescKV) {
	once.Do(func() {
		un_fn = unfn
		logging.LOG.Infof("Start grpc server and listen on %d\n", port)
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

		reg := 0
		if config.GRPC_LB_MODE && nil != fn {
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
				fn(fmt.Sprintf("%s:%d", addr, port))
				reg++
			}
			if 0 == reg {
				logging.LOG.Fatalf(
					"Please using available address \n")
			}
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
		for _, entity := range entities {
			_grpc.RegisterService(&entity.desc, entity.inter)
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
	logging.LOG.Infof("Stop the grpc server...\n")
	if nil != _grpc {
		_grpc.Stop()
	}
	if nil != un_fn && config.GRPC_LB_MODE {
		un_fn()
	}
}
