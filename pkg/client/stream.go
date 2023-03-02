package robocat

import (
	"errors"
	"sync"
)

type RobocatStream[T any] struct {
	channel      chan T
	closed       bool
	waitForWrite sync.WaitGroup
}

func (s *RobocatStream[T]) ensureChannel() chan T {
	if s.channel == nil {
		s.channel = make(chan T)
	}

	return s.channel
}

// Append a new item to the stream.
func (s *RobocatStream[T]) Push(item T) error {
	if s.closed {
		return errors.New("stream channel is closed")
	}

	s.waitForWrite.Add(1)
	s.ensureChannel() <- item
	s.waitForWrite.Done()

	return nil
}

func (s *RobocatStream[T]) Watch(callback func(item T)) {
	for {
		item, ok := <-s.Channel()
		if !ok {
			break
		}

		callback(item)
	}
}

// Get read-only channel with stream items.
func (s *RobocatStream[T]) Channel() <-chan T {
	return s.ensureChannel()
}

// Mark stream as closed.
func (s *RobocatStream[T]) Close() {
	s.closed = true
	s.waitForWrite.Wait()
	close(s.ensureChannel())
}
