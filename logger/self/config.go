package self

import (
	"os"

	"github.com/aserto-dev/self-decision-logger/scribe"
	"github.com/aserto-dev/self-decision-logger/shipper"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type Config struct {
	Port           int            `json:"port"`
	StoreDirectory string         `json:"store_directory"`
	Shipper        shipper.Config `json:"shipper"`
	Scribe         scribe.Config  `json:"scribe"`
}

func (cfg *Config) SetDefaults() {
	if cfg.Port == 0 {
		cfg.Port = 4222
	}

	if cfg.StoreDirectory == "" {
		base, err := os.Getwd()
		if err != nil {
			base = "."
		}

		cfg.StoreDirectory = base + "/nats_store"
	}

	cfg.Shipper.SetDefaults()
}

func mapConfig(cfg map[string]interface{}) (*Config, error) {
	selfCfg := Config{}

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &selfCfg,
		TagName: "json",
	})
	if err != nil {
		return nil, errors.Wrap(err, "error decoding self decision logger config")
	}

	if err := dec.Decode(cfg); err != nil {
		return nil, errors.Wrap(err, "error decoding self decision logger config")
	}

	return &selfCfg, nil
}
