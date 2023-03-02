package robocat

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robocat-ai/robocat/internal/ws"
)

type FlowCommandChain struct {
	client  *Client
	args    *ws.RunnerArguments
	timeout time.Duration
}

type RobocatFlow struct {
	client *Client
	ref    string
	ctx    context.Context
	err    error
}

type RobocatFlowLog struct {
}

func (c *Client) Flow(flow string) *FlowCommandChain {
	return &FlowCommandChain{
		client: c,
		args: &ws.RunnerArguments{
			Flow: flow,
		},
		timeout: 5 * time.Minute,
	}
}

func (chain *FlowCommandChain) WithData(data string) *FlowCommandChain {
	chain.args.Data = data
	return chain
}

func (chain *FlowCommandChain) WithProxy(proxy string) *FlowCommandChain {
	chain.args.Proxy = proxy
	return chain
}

func (chain *FlowCommandChain) WithTimeout(timeout time.Duration) *FlowCommandChain {
	chain.timeout = timeout
	return chain
}

func (f *RobocatFlow) Err() error {
	if f.err != nil {
		return f.err
	}

	return f.client.err
}

func (f *RobocatFlow) Close() {
	f.client.unsubscribe(f.ref)
}

func (f *RobocatFlow) Done() <-chan struct{} {
	return f.ctx.Done()
}

func (f *RobocatFlow) Wait() error {
	for range f.Done() {
		// Wait for context to finish.
	}

	f.Close()

	return f.Err()
}

func (chain *FlowCommandChain) Run() *RobocatFlow {
	flow := &RobocatFlow{
		client: chain.client,
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
		if m.Name == "log" {
			// Redirect log

			if strings.HasPrefix(m.MustText(), "ERROR - ") {
				flow.err = fmt.Errorf("got error log: %v", m.MustText())
				cancel()
			}
		} else if m.Name == "error" {
			flow.err = fmt.Errorf("got error during flow execution: %v", m.MustText())
			cancel()
		} else if m.Name == "output" {
			file, err := ws.ParseFile(m)
			if err != nil {
				flow.err = err
				cancel()
			}

			log.Println("output:", file.Path, file.MimeType, len(file.Payload))
		}
	})

	go func() {
		for range ctx.Done() {
			//
		}

		if ctx.Err() == context.DeadlineExceeded {
			flow.err = context.DeadlineExceeded
		}
	}()

	return flow
}
