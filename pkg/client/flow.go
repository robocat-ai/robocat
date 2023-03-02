package robocat

import (
	"context"
	"fmt"

	"github.com/robocat-ai/robocat/internal/ws"
)

type RobocatFlow struct {
	client *Client
	ref    string
	ctx    context.Context
	err    error

	log    *RobocatLogStream
}

func (chain *FlowCommandChain) Run() *RobocatFlow {
	flow := &RobocatFlow{
		client: chain.client,
		log:    &RobocatLogStream{},
	}

	ref, err := chain.client.sendCommand("run", chain.args)
	if err != nil {
		flow.err = err
		return flow
	}

	ctx, cancel := context.WithTimeout(chain.client.ctx, chain.timeout)

	flow.ref = ref
	flow.ctx = ctx

	chain.client.subscribe(flow.ref, func(ctx context.Context, m *ws.Message) {
		if m.Name == "status" {
			if m.MustText() == "success" {
				cancel()
			}
		} else if m.Name == "log" {
			flow.log.put(m.MustText())
		} else if m.Name == "error" {
			flow.err = errors.New(m.MustText())
			cancel()
		} else if m.Name == "output" {
			if err != nil {
				flow.err = err
				cancel()
			}
		}
	})

	go func() {
		for range ctx.Done() {
			//
		}

		if ctx.Err() == context.DeadlineExceeded {
			flow.err = context.DeadlineExceeded
		}

        flow.Close()
	}()

	return flow
}

func (f *RobocatFlow) Err() error {
	if f.err != nil {
		return f.err
	}

	return f.client.err
}

func (f *RobocatFlow) Close() {
	f.client.unsubscribe(f.ref)
	f.log.Close()
}

func (f *RobocatFlow) Done() <-chan struct{} {
	return f.ctx.Done()
}

func (f *RobocatFlow) Wait() error {
	for range f.Done() {
		// Wait for context to finish.
	}

	return f.Err()
}

func (f *RobocatFlow) Log() *RobocatLogStream {
	return f.log
}
