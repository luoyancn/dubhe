package config

import "time"

var (
	ETCD_ENDPOINTS          = []string{"http://localhost:2379"}
	ETCD_CONNECTION_TIMEOUT = 5 * time.Second
	ETCD_TTL                = 10 * time.Second

	ETCD_SERVICE_NAME = ""
	ETCD_REGISTER_DIR = "/"

	ETCD_USE_TLS  = false
	ETCD_CA_CERT  = "ca.pem"
	ETCD_CA_FILE  = "cert.pem"
	ETCD_KEY_FILE = "key.pem"
)
