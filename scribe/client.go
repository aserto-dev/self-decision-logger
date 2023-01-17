package scribe

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	aserto_client "github.com/aserto-dev/aserto-grpc/grpcclient"
	"github.com/aserto-dev/go-aserto-net/scribe"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientFactory func() (*scribe.Client, error)

func NewClientFactory(ctx context.Context, cfg *Config, dop aserto_client.DialOptionsProvider) ClientFactory {
	return func() (*scribe.Client, error) {
		var conn grpc.ClientConnInterface
		var err error

		if cfg.DisableTLS {
			conn, err = grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, errors.Wrap(err, "error dialing server")
			}
		} else {
			options, err := cfg.Config.ToClientOptions(dop)
			if err != nil {
				return nil, errors.Wrap(err, "error calculating connection options")
			}
			cliConn, err := client.NewConnection(ctx, options...)
			if err != nil {
				return nil, errors.Wrap(err, "error calculating connection options")
			}
			conn = cliConn.Conn
		}

		scribeCli, err := scribe.NewClient(ctx, conn, scribe.AckWaitSeconds(cfg.AckWaitSeconds))
		if err != nil {
			return nil, errors.Wrap(err, "error creating scribe client")
		}

		return scribeCli, nil
	}
}
