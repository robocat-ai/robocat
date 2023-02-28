package robocat

import (
	"context"
	"errors"
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

// Pops the oldest error from the channel and returns it.
// Returns nil if there was no error.
func (f *RobocatFlow) Err() error {
	return f.err
}

func (f *RobocatFlow) Close() {
	f.client.unsubscribe(f.ref)
}

func (f *RobocatFlow) Done() <-chan struct{} {
	return f.ctx.Done()
}

func (f *RobocatFlow) Wait() {
	for range f.Done() {
		// Wait for context to finish.
	}
}

func (chain *FlowCommandChain) Run() *RobocatFlow {
	flow := &RobocatFlow{
		client: chain.client,
		// err: make(chan error),
	}

	// log.Println(chain.args)

	ref, err := chain.client.sendCommand("run", chain.args)
	if err != nil {
		flow.err = err
		return flow
	}

	ctx, cancel := context.WithTimeout(context.Background(), chain.timeout)

	flow.ref = ref
	flow.ctx = ctx

	chain.client.subscribe(flow.ref, func(ctx context.Context, m *ws.Message) {
		if m.Name == "status" && m.MustText() == "ok" {
			//
		} else if m.Name == "log" {
			// Redirect log

			if strings.Contains(strings.ToLower(m.MustText()), "error") {
				flow.err = errors.New("got error log")
				cancel()
			}

			log.Println(flow, m.MustText())
		}
	})

	go func() {
		for range ctx.Done() {
		}

		log.Println(ctx.Err())

		if ctx.Err() == context.DeadlineExceeded {
			flow.err = context.DeadlineExceeded
		}
	}()

	// time.Sleep(time.Second * 5)

	// loop:
	// 	for {
	// 		select {
	// 		case <-time.After(time.Second * 5):
	// 			break loop
	// 		default:
	// 			msg, err := chain.client.readUpdate()
	// 			log.Println(msg, err)
	// 		}
	// 	}

	return flow
}
