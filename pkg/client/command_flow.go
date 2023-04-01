package robocat

import (
	"context"
	"errors"
	"time"

	"github.com/robocat-ai/robocat/internal/ws"
)

type FlowCommandChain struct {
	client  *Client
	args    *ws.RunnerArguments
	timeout time.Duration
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

func (chain *FlowCommandChain) Run() *RobocatFlow {
	flow := &RobocatFlow{
		client: chain.client,
		log:    &RobocatLogStream{},
		output: &RobocatFileStream{},
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
			flow.log.Push(m.MustText())
		} else if m.Name == "error" {
			flow.err = errors.New(m.MustText())
			cancel()
		} else if m.Name == "output" {
			file, err := ws.ParseFileFromMessage(m)
			if err != nil {
				flow.err = err
				cancel()
			}

			flow.output.Push(&File{
				Path:     file.Path,
				MimeType: file.MimeType,
				Payload:  file.Payload,
			})
		}
	})

	go func() {
		defer flow.close()
		defer cancel()

		flow.errWait.Add(1)
		defer flow.errWait.Done()

		for {
			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					flow.err = context.DeadlineExceeded
				}
				return
			case <-chain.client.cancelFlowChannel():
				flow.err = errors.New("flow was aborted")
				return
			}
		}
	}()

	return flow
}
