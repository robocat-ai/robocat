package robocat

import (
	"context"
)

type RobocatFlow struct {
	client *Client
	ref    string
	ctx    context.Context
	err    error

	log    *RobocatLogStream
	output *RobocatFileStream
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
	f.output.Close()
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

func (f *RobocatFlow) Files() *RobocatFileStream {
	return f.output
}
