package scribe

import (
	"context"

	client "github.com/aserto-dev/go-aserto"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type ClientFactory func() (*Client, error)

func NewClientFactory(ctx context.Context, cfg *Config, dopts ...grpc.DialOption) ClientFactory {
	if cfg.DisableTLS {
		cfg.NoTLS = true
	}

	return func() (*Client, error) {
		conn, err := cfg.Config.Connect(client.WithDialOptions(dopts...))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create grpc client")
		}

		scribeCli, err := NewClient(ctx, conn, AckWaitSeconds(cfg.AckWaitSeconds))
		if err != nil {
			return nil, errors.Wrap(err, "error creating scribe client")
		}

		return scribeCli, nil
	}
}
