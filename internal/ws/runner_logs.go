package ws

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"time"
)

func (r *RobocatRunner) watchLogs(
	ctx context.Context,
	message *Message,
	stream io.Reader,
) {
	log.Debugw("Watching logs", "ref", message.Ref)

	scanner := bufio.NewScanner(stream)

	startPrefix := "START - automation started"
	errorPrefix := "ERROR - "

	var automationWatchdogTimer *time.Timer
	automationWatchdogTimeout, err := time.ParseDuration(os.Getenv("AUTOMATION_START_TIMEOUT"))
	if err != nil {
		automationWatchdogTimeout = time.Minute
	}

loop:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			// Stop logging when parent context is done.
			break loop
		default:
			line := scanner.Text()
			message.Reply("log", scanner.Text())

			if strings.HasPrefix(line, startPrefix) {
				automationWatchdogTimer = time.AfterFunc(
					automationWatchdogTimeout, func() {
						r.cancel()
						message.ReplyWithErrorf(
							"automation start timeout reached (%s)",
							automationWatchdogTimeout,
						)
					},
				)
			} else if strings.HasPrefix(line, errorPrefix) {
				r.cancel()
				message.ReplyWithErrorf(
					"got error during run execution: %s",
					strings.TrimPrefix(line, errorPrefix),
				)
				break loop
			} else {
				if automationWatchdogTimer != nil {
					automationWatchdogTimer.Stop()
				}
			}
		}
	}

	if automationWatchdogTimer != nil {
		automationWatchdogTimer.Stop()
	}

	log.Debugw("Stopped watching logs", "ref", message.Ref)
}
