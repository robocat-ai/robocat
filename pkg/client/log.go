package robocat

import (
	"errors"
	"sync"
)

type RobocatLog struct {
	output       chan string
	closed       bool
	waitForWrite sync.WaitGroup
}

func (l *RobocatLog) ensureOutput() chan string {
	if l.output == nil {
		l.output = make(chan string)
	}

	return l.output
}

// Append a new line to the log stream.
// This method is used internally by the
// RobocatFlow to append lines to the stream.
func (l *RobocatLog) append(line string) error {
	if l.closed {
		return errors.New("log channel is closed")
	}

	l.waitForWrite.Add(1)
	l.ensureOutput() <- line
	l.waitForWrite.Done()

	return nil
}

// Get read-only channel to which log lines
// are being put.
func (l *RobocatLog) Channel() <-chan string {
	return l.ensureOutput()
}

// Mark log stream as closed.
func (l *RobocatLog) Close() {
	l.closed = true
	l.waitForWrite.Wait()
	close(l.ensureOutput())
}
