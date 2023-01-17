package scribe

import "github.com/aserto-dev/aserto-grpc/grpcclient"

type Config struct {
	grpcclient.Config `json:",squash"` // nolint:staticcheck // squash is used by mapstructure

	MaxInflightBatches int  `json:"max_inflight_batches"`
	AckWaitSeconds     int  `json:"ack_wait_seconds"`
	DisableTLS         bool `json:"disable_tls"`
}
