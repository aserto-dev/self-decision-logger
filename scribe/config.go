package scribe

import "github.com/aserto-dev/go-aserto/client"

type Config struct {
	client.Config `json:",squash"` // nolint:staticcheck // squash is used by mapstructure

	MaxInflightBatches int  `json:"max_inflight_batches"`
	AckWaitSeconds     int  `json:"ack_wait_seconds"`
	DisableTLS         bool `json:"disable_tls"`
}
