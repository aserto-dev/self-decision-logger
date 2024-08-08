package scribe

import (
	"context"

	client "github.com/aserto-dev/go-aserto"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientFactory func() (*Client, error)

func NewClientFactory(ctx context.Context, cfg *Config, dop client.DialOptionsProvider) ClientFactory {
	return func() (*Client, error) {
		var conn grpc.ClientConnInterface
		var err error

		if cfg.DisableTLS {
			conn, err = grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, errors.Wrap(err, "error dialing server")
			}
		} else {
			options, err := cfg.Config.ToConnectionOptions(dop)
			if err != nil {
				return nil, errors.Wrap(err, "error calculating connection options")
			}
			cliConn, err := client.NewConnection(options...)
			if err != nil {
				return nil, errors.Wrap(err, "error calculating connection options")
			}
			conn = cliConn
		}

		scribeCli, err := NewClient(ctx, conn, AckWaitSeconds(cfg.AckWaitSeconds))
		if err != nil {
			return nil, errors.Wrap(err, "error creating scribe client")
		}

		return scribeCli, nil
	}
}
