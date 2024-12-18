package scribe

import (
	"context"

	client "github.com/aserto-dev/go-aserto"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type ClientFactory func() (*Client, error)

func NewClientFactory(ctx context.Context, cfg *Config, dopts ...grpc.DialOption) (ClientFactory, error) {
	if cfg.DisableTLS {
		// Backwards compatibility
		cfg.NoTLS = true
	}

	conn, err := cfg.Config.Connect(client.WithDialOptions(dopts...))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create grpc connection")
	}

	return func() (*Client, error) {
		scribeCli, err := NewClient(ctx, conn, AckWaitSeconds(cfg.AckWaitSeconds))
		return scribeCli, errors.Wrap(err, "error creating scribe client")
	}, nil
}
