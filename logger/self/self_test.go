package self_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/aserto-dev/aserto-grpc/grpcclient"
	"github.com/aserto-dev/go-aserto-net/scribe"
	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	scribe_grpc "github.com/aserto-dev/go-grpc/aserto/scribe/v2"
	scribe_cli "github.com/aserto-dev/self-decision-logger/scribe"
	"github.com/aserto-dev/self-decision-logger/shipper"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/self-decision-logger/logger/self"
)

const (
	scribeAddress = "localhost:9191"
)

var cfg = self.Config{
	Port:           4222,
	StoreDirectory: "./nats_store",
	Shipper: shipper.Config{
		DeleteStreamOnDone:    true,
		PublishTimeoutSeconds: 1,
	},
	Scribe: scribe_cli.Config{
		Config: grpcclient.Config{
			Address: scribeAddress,
		},
		AckWaitSeconds: 10,
		DisableTLS:     true,
	},
}

func startScribe(ctx context.Context, assert *require.Assertions, cb scribe.ServerBatchFunc) func() {
	l, err := net.Listen("tcp", scribeAddress)
	if err != nil {
		assert.FailNow(err.Error())
	}

	grpcSrv := grpc.NewServer()
	scribeSrv, cleanup, err := scribe.NewServer(ctx, cb)
	if err != nil {
		assert.FailNow(err.Error())
	}

	grpcSrv.RegisterService(&scribe_grpc.Scribe_ServiceDesc, scribeSrv)
	go func() {
		err := grpcSrv.Serve(l)
		if err != nil {
			assert.FailNow("server failed to start")
		}
	}()

	return func() {
		cleanup()
		grpcSrv.Stop()
	}
}

func runServer(ctx context.Context, assert *require.Assertions, logs int, done chan<- bool, received map[string]bool) func() {
	recv := make(chan *scribe.Batch)

	cleanup := startScribe(ctx, assert, func(ctx context.Context, b *scribe.Batch) {
		recv <- b
	})

	count := logs
	go func() {
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
		done <- true
	}()

	return cleanup
}

func makeDecision() *api.Decision {
	return &api.Decision{
		Id: uuid.NewString(),
		// TenantId:  "e5e07c3c-c449-11eb-a518-0045ec92c058",
		Timestamp: timestamppb.New(time.Date(2021, time.September, 02, 17, 22, 0, 0, time.UTC)),
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

	dlog, err := self.NewFromConfig(ctx, &cfg, &l, grpcclient.NewDialOptionsProvider())
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

	dlog, err := self.NewFromConfig(ctx, &cfg, &l, grpcclient.NewDialOptionsProvider())
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
