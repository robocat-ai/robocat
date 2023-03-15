package robocat

import (
	"context"
	"fmt"
	"sync"
)

type RobocatFlow struct {
	client *Client
	ref    string
	ctx    context.Context
	err    error

	errWait sync.WaitGroup

	log    *RobocatLogStream
	output *RobocatFileStream
}

func (f *RobocatFlow) Err() error {
	if f.err != nil {
		return fmt.Errorf("flow error: %w", f.err)
	}

	if f.client.err != nil {
		return fmt.Errorf("client error: %w", f.client.err)
	}

	return nil
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

	f.errWait.Wait()
	return f.Err()
}

func (f *RobocatFlow) Log() *RobocatLogStream {
	return f.log
}

func (f *RobocatFlow) Files() *RobocatFileStream {
	return f.output
}
