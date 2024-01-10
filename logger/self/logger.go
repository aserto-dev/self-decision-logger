package self

import (
	"context"
	"fmt"
	"time"

	"github.com/aserto-dev/go-aserto/client"
	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/self-decision-logger/scribe"
	"github.com/aserto-dev/self-decision-logger/shipper"
	decisionlog "github.com/aserto-dev/topaz/decision_log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	nats_server "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	stream  = "decision-log-stream-v2"
	subject = "decision-logs-v2"
)

type selfLogger struct {
	jsCtx      nats.JetStreamContext
	natsServer *nats_server.Server
	shipper    *shipper.Shipper
}

func New(ctx context.Context, cfg map[string]interface{}, logger *zerolog.Logger, dop client.DialOptionsProvider) (decisionlog.DecisionLogger, error) {
	selfCfg, err := mapConfig(cfg)
	if err != nil {
		return nil, err
	}

	return NewFromConfig(ctx, selfCfg, logger, dop)
}

func NewFromConfig(ctx context.Context, cfg *Config, logger *zerolog.Logger, dop client.DialOptionsProvider) (decisionlog.DecisionLogger, error) {
	cfg.SetDefaults()

	opts := &nats_server.Options{
		Host:      "localhost",
		Port:      cfg.Port,
		JetStream: true,
		StoreDir:  cfg.StoreDirectory,
	}
	natsServer, err := nats_server.NewServer(opts)
	if err != nil {
		return nil, errors.Wrap(err, "error starting nats server")
	}
	go natsServer.Start()
	natsServer.ReadyForConnections(time.Second * 10)
	natsCli, err := nats.Connect(fmt.Sprintf("localhost:%d", cfg.Port))
	if err != nil {
		logger.Err(err).Msg("error connecting NATS client")
		return nil, err
	}
	if err != nil {
		return nil, errors.Wrap(err, "error creating nats client")
	}

	scf := scribe.NewClientFactory(ctx, &cfg.Scribe, dop)
	if err != nil {
		return nil, errors.Wrap(err, "error creating scribe client")
	}

	shpr, err := shipper.New(ctx, logger, &cfg.Shipper, natsCli, scf)
	if err != nil {
		return nil, errors.Wrap(err, "error creating lumberjack client")
	}
	jsCtx, err := natsCli.JetStream()
	if err != nil {
		return nil, errors.Wrap(err, "error establishing jetstream context")
	}
	l := &selfLogger{
		jsCtx:      jsCtx,
		natsServer: natsServer,
		shipper:    shpr,
	}

	return l, nil
}

func (l *selfLogger) Log(d *api.Decision) error {
	pub, err := anypb.New(d)
	if err != nil {
		return errors.Wrap(err, "error creating any wrapper")
	}

	bytes, err := proto.Marshal(pub)
	if err != nil {
		return errors.Wrap(err, "error marshaling decision")
	}

	_, err = l.jsCtx.Publish(subject, bytes, nats.ExpectStream(stream))
	if err != nil {
		return errors.Wrap(err, "error publishing decision to stream")
	}

	return nil
}

func (l *selfLogger) Shutdown() {
	if l.shipper != nil {
		l.shipper.Shutdown()
	}
	if l.natsServer != nil {
		l.natsServer.Shutdown()
	}
}
