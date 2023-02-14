package shipper

import (
	"context"
	"math"
	"time"

	go_aserto_net_scribe "github.com/aserto-dev/go-aserto-net/scribe"
	"github.com/aserto-dev/self-decision-logger/scribe"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	stream  = "decision-log-stream-v2"
	subject = "decision-logs-v2"
)

type Shipper struct {
	ctx       context.Context
	logger    *zerolog.Logger
	cancel    context.CancelFunc
	cfg       *Config
	jsCtx     nats.JetStreamContext
	subs      *nats.Subscription
	scf       scribe.ClientFactory
	scribeCli *go_aserto_net_scribe.Client
	queueCh   chan *nats.Msg
	timer     *time.Timer
	batch     []*nats.Msg
}

func New(ctx context.Context, logger *zerolog.Logger, cfg *Config, natsCli *nats.Conn, scf scribe.ClientFactory) (*Shipper, error) {
	cfg.SetDefaults()

	newLogger := logger.With().Str("component", "shipper").Logger()
	jsCtx, err := natsCli.JetStream()
	if err != nil {
		return nil, errors.Wrap(err, "error establishing jetstream context")
	}
	if cfg.MaxBatchSize < 1 {
		return nil, errors.New("invalid max batch size, must be > 0")
	}
	if cfg.PublishTimeoutSeconds < 1 {
		return nil, errors.New("invalid publish timeout error, must be > 0")
	}
	_, err = jsCtx.AddStream(&nats.StreamConfig{
		Name:        stream,
		Description: "Authorizer decision logs stream",
		Subjects:    []string{subject},
		Retention:   nats.WorkQueuePolicy,
		Storage:     nats.FileStorage,
		MaxBytes:    cfg.MaxBytes,
		Discard:     nats.DiscardNew,
	})
	if err != nil && errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		return nil, errors.Wrap(err, "error adding decision-logs stream")
	}
	ctx, cancel := context.WithCancel(ctx)

	s := &Shipper{
		ctx:     ctx,
		logger:  &newLogger,
		cancel:  cancel,
		cfg:     cfg,
		jsCtx:   jsCtx,
		scf:     scf,
		timer:   time.NewTimer(time.Second * time.Duration(cfg.PublishTimeoutSeconds)),
		queueCh: make(chan *nats.Msg, cfg.MaxBatchSize*cfg.MaxInflightBatches),
	}

	if !s.timer.Stop() {
		<-s.timer.C
	}
	err = s.run()
	if err != nil {
		return nil, err
	}
	return s, nil
}
func (s *Shipper) Shutdown() {
	s.cancel()
	_ = s.subs.Unsubscribe()
	if s.cfg.DeleteStreamOnDone && s.jsCtx != nil {
		_ = s.jsCtx.DeleteStream(stream)
	}
}
func (s *Shipper) run() error {
	subs, err := s.jsCtx.ChanQueueSubscribe(subject, "decision-queue", s.queueCh,
		nats.Context(s.ctx),
		nats.BindStream(stream),
		nats.ManualAck(),
		nats.AckWait(time.Second*time.Duration(s.cfg.AckWaitSeconds)),
		nats.MaxAckPending(cap(s.queueCh)),
		nats.MaxDeliver(math.MaxInt))
	if err != nil {
		return errors.Wrap(err, "error subscribing to decision queue")
	}
	s.subs = subs
	go func() {
		for {
			select {
			case msg := <-s.queueCh:
				s.handleMsg(msg)
			case <-s.timer.C:
				s.publishBatch()
			case <-s.ctx.Done():
				return
			}
		}
	}()
	s.logger.Info().Msg("shipper is running")
	return nil
}
func (s *Shipper) handleMsg(msg *nats.Msg) {
	if s.batch == nil {
		s.timer.Reset(time.Second * time.Duration(s.cfg.PublishTimeoutSeconds))
	}
	s.batch = append(s.batch, msg)
	if len(s.batch) == s.cfg.MaxBatchSize {
		s.publishBatch()
	}
}
func (s *Shipper) publishBatch() {
	s.logger.Trace().Msgf("publishing batch with size: %d", len(s.batch))

	data := make([]*anypb.Any, 0, len(s.batch))
	for _, msg := range s.batch {
		any := anypb.Any{}
		err := proto.Unmarshal(msg.Data, &any)
		if err != nil {
			s.logger.Error().Err(err).Msg("error unmarshalling message")
			s.nak(msg)
			continue
		}
		data = append(data, &any)
	}

	curBatch := s.batch
	s.batch = nil

	cli, err := s.getScribeClient()
	if err != nil {
		s.handlePublishError(err, curBatch)
		return
	}

	err = cli.WriteBatch(s.ctx, data, func(ack bool, err error) {
		var f func(*nats.Msg)
		if err == nil {
			f = s.ack
		} else {
			s.logger.Info().Err(err).Msg("received an error in ack callback")
			f = s.nak
		}
		for _, msg := range curBatch {
			f(msg)
		}
		s.logger.Trace().Msgf("processed %d acks", len(curBatch))
	})

	if err != nil {
		s.handlePublishError(err, curBatch)
		return
	}

	if !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
	s.logger.Trace().Msg("published batch")
}
func (s *Shipper) ack(msg *nats.Msg) {
	_ = msg.Ack(nats.Context(s.ctx))
}
func (s *Shipper) nak(msg *nats.Msg) {
	_ = msg.Nak(nats.Context(s.ctx))
}

func (s *Shipper) getScribeClient() (*go_aserto_net_scribe.Client, error) {
	if s.scribeCli != nil {
		return s.scribeCli, nil
	}

	var err error
	s.scribeCli, err = s.scf()
	if err != nil {
		s.scribeCli = nil
		return nil, err
	}

	s.logger.Trace().Msg("created new scribe client")

	return s.scribeCli, nil
}

func (s *Shipper) handlePublishError(err error, batch []*nats.Msg) {
	s.logger.Info().Err(err).Msg("error publishing batch")
	for _, msg := range batch {
		s.nak(msg)
	}
	s.scribeCli = nil
}
