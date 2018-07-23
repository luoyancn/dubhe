package config

import "time"

var (
	GRPC_PORT                = 8080
	GRPC_USE_TLS             = false
	GRPC_CA_FILE             = ""
	GRPC_KEY_FILE            = ""
	GRPC_LB_MODE             = false
	GRPC_REGISTERED_ADDRESS  = []string{}
	GRPC_DEBUG               = false
	GRPC_CONCURRENCY         = 1024
	GRPC_REQ_MAX_FREQUENCY   = 1024.00
	GRPC_REQ_BURST_FREQUENCY = 10
	GRPC_CONNECTION_LIMIT    = 10240
	GRPC_TIMEOUT             = 10 * time.Second
)
