package scribe

import (
	"context"

	client "github.com/aserto-dev/go-aserto"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientFactory func() (*Client, error)

func NewClientFactory(ctx context.Context, cfg *Config, dopts ...grpc.DialOption) ClientFactory {
	return func() (*Client, error) {
		var conn *grpc.ClientConn
		var err error

		if cfg.DisableTLS {
			conn, err = client.NewConnection(client.WithAddr(cfg.Address), client.WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())))
			if err != nil {
				return nil, errors.Wrap(err, "error dialing server")
			}
		} else {
			if cliConn, err := cfg.Config.Connect(client.WithDialOptions(dopts...)); err != nil {
				conn = cliConn
			}
		}

		scribeCli, err := NewClient(ctx, conn, AckWaitSeconds(cfg.AckWaitSeconds))
		if err != nil {
			return nil, errors.Wrap(err, "error creating scribe client")
		}

		return scribeCli, nil
	}
}
