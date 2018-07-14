package config

import "time"

var (
	ETCD_ENDPOINTS          = []string{"http://localhost:2379"}
	ETCD_CONNECTION_TIMEOUT = 5 * time.Second
	ETCD_TTL                = 10 * time.Second
)
