package ws

import (
	"bufio"
	"context"
	"io"
	"strings"
)

func (r *RobocatRunner) watchLogs(
	ctx context.Context,
	message *Message,
	stream io.Reader,
) {
	log.Debug("Watching logs")

	scanner := bufio.NewScanner(stream)

	errorPrefix := "ERROR - "

loop:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			// Stop logging when parent context is done.
			break loop
		default:
			line := scanner.Text()
			if strings.HasPrefix(line, errorPrefix) {
				r.cancel()
				message.ReplyWithErrorf(
					"got error during run execution: %s",
					strings.TrimPrefix(line, errorPrefix),
				)
				break loop
			} else {
				message.Reply("log", scanner.Text())
			}
		}
	}

	log.Debug("Stopped watching logs")
}
