package scribe

import client "github.com/aserto-dev/go-aserto"

type Config struct {
	client.Config `json:",squash"` //nolint:staticcheck,tagliatelle

	MaxInflightBatches int `json:"max_inflight_batches"`
	AckWaitSeconds     int `json:"ack_wait_seconds"`
	// Deprecated: use NoTLS instead
	DisableTLS bool `json:"disable_tls"`
}
