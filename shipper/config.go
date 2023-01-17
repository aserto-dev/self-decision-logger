package shipper

type Config struct {
	MaxBytes              int64 `json:"max_bytes"`
	MaxBatchSize          int   `json:"max_batch_size"`
	PublishTimeoutSeconds int   `json:"publish_timeout_seconds"`
	MaxInflightBatches    int   `json:"max_inflight_batches"`
	AckWaitSeconds        int   `json:"ack_wait_seconds"`
	DeleteStreamOnDone    bool  `json:"delete_stream_on_done"`
	BackoffSeconds        []int `json:"backoff_seconds"`
}

var (
	defaultBackoff = []int{5, 10, 30, 60, 120, 300}
)

func (cfg *Config) SetDefaults() {
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = 100 * 1024 * 1024
	}
	if cfg.MaxBatchSize == 0 {
		cfg.MaxBatchSize = 512
	}
	if cfg.PublishTimeoutSeconds == 0 {
		cfg.PublishTimeoutSeconds = 10
	}
	if cfg.MaxInflightBatches == 0 {
		cfg.MaxInflightBatches = 10
	}
	if cfg.AckWaitSeconds == 0 {
		cfg.AckWaitSeconds = 60
	}
	if cfg.BackoffSeconds == nil {
		cfg.BackoffSeconds = defaultBackoff
	}
}
