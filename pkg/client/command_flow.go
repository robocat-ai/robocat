package robocat

import (
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
