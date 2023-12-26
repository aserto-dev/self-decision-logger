package scribe

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	scribe "github.com/aserto-dev/go-decision-logs/aserto/scribe/v2"
	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

type ackWaiter struct {
	cb    ClientBatchFunc
	timer *time.Timer
}

type clientOptions struct {
	maxInflight    int
	ackWaitSeconds int
}

type ClientOpt interface {
	applyClient(opts *clientOptions) error
}

type AckWaitSeconds int

var (
	ErrAck        = errors.New("ack wait must be > 0")
	ErrAckTimeout = errors.New("timed out waiting for acknowledgement")
)

func (w AckWaitSeconds) applyClient(cliOpts *clientOptions) error {
	if w < 1 {
		return ErrAck
	}

	cliOpts.ackWaitSeconds = int(w)
	return nil
}

type Client struct {
	id        string
	nextBatch uint64
	wbCli     scribe.Scribe_WriteBatchClient
	inflight  map[string]ackWaiter
	mtx       sync.Mutex
	sem       *semaphore.Weighted
	ackWait   time.Duration
}

type ClientBatchFunc func(ack bool, err error)

func NewClient(ctx context.Context, conn grpc.ClientConnInterface, opts ...ClientOpt) (*Client, error) {
	cliOpt := clientOptions{
		maxInflight:    20,
		ackWaitSeconds: 10,
	}

	for _, o := range opts {
		err := o.applyClient(&cliOpt)
		if err != nil {
			return nil, err
		}
	}

	cli := scribe.NewScribeClient(conn)
	wbCli, err := cli.WriteBatch(ctx)
	if err != nil {
		return nil, err
	}

	c := &Client{
		id:       uuid.New().String(),
		wbCli:    wbCli,
		inflight: map[string]ackWaiter{},
		sem:      semaphore.NewWeighted(int64(cliOpt.maxInflight)),
		ackWait:  time.Second * time.Duration(cliOpt.ackWaitSeconds),
	}
	go c.ackLoop()

	return c, nil
}

func (c *Client) WriteBatch(ctx context.Context, batch []*anypb.Any, cb ClientBatchFunc) error {
	err := c.sem.Acquire(ctx, 1)
	if err != nil {
		return err
	}

	id := c.addWaiter(cb)

	err = c.wbCli.Send(&scribe.WriteBatchRequest{
		Id:    id,
		Batch: batch,
	})

	if err != nil {
		c.deleteWaiter(id)
		return err
	}

	return nil
}

func (c *Client) ackLoop() {
	for {
		resp, err := c.wbCli.Recv()
		if err != nil {
			_ = c.wbCli.CloseSend()
			break
		}

		w := c.deleteWaiter(resp.Id)
		if w != nil {
			w.cb(resp.Ack, nil)
		}
	}
}

func (c *Client) ackTimeout(id string) {
	w := c.deleteWaiter(id)
	if w != nil {
		w.cb(false, ErrAckTimeout)
	}
}

func (c *Client) addWaiter(cb ClientBatchFunc) string {
	c.mtx.Lock()
	c.nextBatch++
	id := fmt.Sprintf("%s:%d", c.id, c.nextBatch)
	c.inflight[id] = ackWaiter{
		cb: cb,
		timer: time.AfterFunc(c.ackWait, func() {
			c.ackTimeout(id)
		}),
	}
	c.mtx.Unlock()
	return id
}

func (c *Client) deleteWaiter(id string) *ackWaiter {
	c.mtx.Lock()
	w, ok := c.inflight[id]
	delete(c.inflight, id)
	c.mtx.Unlock()

	if !ok {
		return nil
	}

	c.sem.Release(1)
	_ = w.timer.Stop()
	return &w
}
