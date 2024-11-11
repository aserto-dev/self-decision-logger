package self_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	scribe_grpc "github.com/aserto-dev/go-decision-logs/aserto/scribe/v2"
	self "github.com/aserto-dev/self-decision-logger/logger/self"
	scribe_cli "github.com/aserto-dev/self-decision-logger/scribe"
	shipper "github.com/aserto-dev/self-decision-logger/shipper"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	scribeAddress = "localhost:9191"
)

var cfg = self.Config{
	Port:           4222,
	StoreDirectory: "./nats_store",
	Shipper: shipper.Config{
		MaxBatchSize:          500, // must be a divisor of logsPerServer.
		DeleteStreamOnDone:    true,
		PublishTimeoutSeconds: 1,
	},
	Scribe: scribe_cli.Config{
		Config: client.Config{
			Address: scribeAddress,
		},
		AckWaitSeconds: 10,
		DisableTLS:     true,
	},
}

type mockBatch struct {
	ID    string
	wbSrv scribe_grpc.Scribe_WriteBatchServer
	Batch []*anypb.Any
}

func (m *mockBatch) Ack() {
	err := m.wbSrv.Send(&scribe_grpc.WriteBatchResponse{
		Id:  m.ID,
		Ack: true,
	})
	if err != nil {
		panic(err)
	}
}

type mockServer struct {
	ctx     context.Context
	cancel  context.CancelFunc
	errCh   chan error
	handler func(ctx context.Context, batch *mockBatch)
}

func newMockServer(ctx context.Context, cb func(ctx context.Context, batch *mockBatch)) *mockServer {
	cctx, cancel := context.WithCancel(ctx)
	return &mockServer{
		ctx:     cctx,
		cancel:  cancel,
		handler: cb,
	}
}

func (s *mockServer) WriteBatch(wbs scribe_grpc.Scribe_WriteBatchServer) error {
	defer s.cancel()

	go func() {
		for {
			req, err := wbs.Recv()
			if err == io.EOF {
				s.errCh <- nil
				break
			}
			if err != nil {
				s.errCh <- err
				break
			}

			b := mockBatch{
				ID:    req.Id,
				wbSrv: wbs,
				Batch: req.Batch,
			}

			go s.handler(wbs.Context(), &b)
		}
	}()

	select {
	case <-wbs.Context().Done():
		return nil
	case <-s.ctx.Done():
		return nil
	case err := <-s.errCh:
		return err
	}
}

func startScribe(ctx context.Context, assert *require.Assertions, cb func(ctx context.Context, batch *mockBatch)) func() {
	l, err := net.Listen("tcp", scribeAddress)
	if err != nil {
		assert.FailNow(err.Error())
	}

	grpcSrv := grpc.NewServer()

	scribeSrv := newMockServer(ctx, cb)

	grpcSrv.RegisterService(&scribe_grpc.Scribe_ServiceDesc, scribeSrv)
	go func() {
		err := grpcSrv.Serve(l)
		if err != nil {
			assert.FailNow("server failed to start")
		}
	}()

	return func() {
		grpcSrv.Stop()
	}
}

func runServer(ctx context.Context, assert *require.Assertions, logs int, done chan<- bool, received map[string]bool) func() {
	recv := make(chan *mockBatch)

	cleanup := startScribe(ctx, assert, func(_ context.Context, b *mockBatch) {
		recv <- b
	})

	go func() {
		count := logs
		for count > 0 {
			select {
			case b := <-recv:
				for _, any := range b.Batch {
					d := api.Decision{}
					err := anypb.UnmarshalTo(any, &d, proto.UnmarshalOptions{})
					if err != nil {
						assert.FailNow("failed to unmarshal decision")
						count = -1
						continue
					}
					received[d.Id] = true
				}
				count -= len(b.Batch)
				b.Ack()
			case <-ctx.Done():
				count = -1
			}
		}
		assert.Zero(count)
		done <- true
	}()

	return cleanup
}

func makeDecision() *api.Decision {
	return &api.Decision{
		Id: uuid.NewString(),
		// TenantId:  "e5e07c3c-c449-11eb-a518-0045ec92c058",
		Timestamp: timestamppb.New(time.Date(2021, time.September, 2, 17, 22, 0, 0, time.UTC)),
		User: &api.DecisionUser{
			Context: &api.IdentityContext{
				Identity: "some@name.org",
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			},
			Id:    "011a88bc-7df9-4d92-ba1f-2ff319e101e1",
			Email: "some@name.org",
		},
		Policy: &api.DecisionPolicy{
			Context: &api.PolicyContext{
				Path:      "some/policy/path/here",
				Decisions: []string{"read", "write"},
			},
			PolicyInstance: &api.PolicyInstance{
				Name:          "test",
				InstanceLabel: "test",
			},
			RegistryService: "registry.test",
			RegistryImage:   "mypolicy",
			RegistryTag:     "1.0.1",
			RegistryDigest:  "adigest1234",
		},
		Outcomes: map[string]bool{
			"read":  true,
			"write": false,
		},
	}
}

func TestSelfLogger(t *testing.T) {
	assert := require.New(t)
	l := zerolog.Nop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	received := map[string]bool{}
	done := make(chan bool)
	logs := 10000

	cleanup := runServer(ctx, assert, logs, done, received)
	defer cleanup()

	dlog, err := self.NewFromConfig(ctx, &cfg, &l)
	assert.NoError(err)
	defer dlog.Shutdown()

	start := time.Now()

	for i := 0; i < logs; i++ {
		err := dlog.Log(makeDecision())
		assert.NoError(err)
	}

	select {
	case <-done:
		fmt.Printf("elapsed: %s\n", time.Since(start))
	case <-time.After(time.Second * 30):
		assert.Fail("timed out")
	}
	assert.Equal(logs, len(received))
}

func TestSelfLoggerWithDisconnect(t *testing.T) {
	assert := require.New(t)
	l := zerolog.Nop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	received := map[string]bool{}
	done := make(chan bool)
	logs := 50000
	logsPerServer := 10000

	go func() {
		for i := 0; i < 5; i++ {
			serverDone := make(chan bool)
			cleanup := runServer(ctx, assert, logsPerServer, serverDone, received)
			select {
			case <-serverDone:
			case <-time.After(time.Second * 30):
				assert.Fail("timed out")
			}
			cleanup()
		}

		done <- true
	}()

	dlog, err := self.NewFromConfig(ctx, &cfg, &l)
	assert.NoError(err)
	defer dlog.Shutdown()

	start := time.Now()

	for i := 0; i < logs; i++ {
		err := dlog.Log(makeDecision())
		assert.NoError(err)
	}

	select {
	case <-done:
		fmt.Printf("elapsed: %s\n", time.Since(start))
	case <-time.After(time.Second * 30):
		assert.Fail("timed out")
	}
	assert.Equal(logs, len(received))
}
