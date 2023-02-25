package ws

import (
	"bufio"
	"context"
	"io"
)

func (r *RobocatRunner) watchLogs(
	ctx context.Context,
	message *Message,
	stream io.Reader,
) {
	log.Debug("Watching logs")

	scanner := bufio.NewScanner(stream)

loop:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			// Stop logging when parent context is done.
			break loop
		default:
			message.Reply("log", scanner.Text())
		}
	}

	log.Debug("Stopped watching logs")
}
