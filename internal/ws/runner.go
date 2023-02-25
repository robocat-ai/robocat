package ws

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"time"
)

type RunnerArguments struct {
	Flow  string `json:"flow"`
	Data  string `json:"data"`
	Proxy string `json:"proxy"`
}

func (a *RunnerArguments) ToArray() []string {
	args := []string{a.Flow}

	if len(a.Data) > 0 {
		args = append(args, "--data", a.Data)
	}

	if len(a.Proxy) > 0 {
		u, err := url.Parse(a.Proxy)
		if err == nil {
			address := fmt.Sprintf("%s:%s", u.Hostname(), u.Port())
			if u.User != nil {
				address = fmt.Sprintf("%s@%s", u.User.String(), address)
			}

			args = append(args, "--proxy-protocol", u.Scheme)
			args = append(args, "--proxy-address", address)
		}
	}

	return args
}

type RobocatRunner struct {
	abortScheduledCleanupSignal chan bool
	cleanupScheduled            bool
}

func NewRobocatRunner() *RobocatRunner {
	return &RobocatRunner{
		abortScheduledCleanupSignal: make(chan bool),
		cleanupScheduled:            false,
	}
}

func (r *RobocatRunner) scheduleCleanup() {
	r.cleanupScheduled = true

	for {
		select {
		case <-r.abortScheduledCleanupSignal:
			log.Debug("Scheduled clean-up aborted")
			r.cleanupScheduled = false
			return
		case <-time.After(time.Second * 5):
			log.Debug("Cleaning up previous TagUI session")
			r.cleanup()
			r.cleanupScheduled = false
			return
		}
	}
}

func (r *RobocatRunner) abortScheduledCleanup() {
	if r.cleanupScheduled {
		r.abortScheduledCleanupSignal <- true
	}
}

func (r *RobocatRunner) cleanup() {
	exec.Command("kill_tagui").Run()
}

func (r *RobocatRunner) streamLogs(
	ctx context.Context,
	server *Server,
	stream io.Reader,
) {
	scanner := bufio.NewScanner(stream)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			// Stop logging when parent context is done.
			return
		default:
			server.Send("log", scanner.Text())
		}
	}
}

func (r *RobocatRunner) Handle(
	ctx context.Context,
	server *Server,
	message *Message,
) {
	// In case of quick disconnect right after connection TagUI flow can
	// still be running, so we try to kill previously running TagUI instance
	// using scheduled clean-up. However, when the flow is run again we must
	// abort scheduled clean-up and instead clean-up manually.
	r.abortScheduledCleanup()
	r.cleanup()

	var args *RunnerArguments

	err := json.Unmarshal(message.Body, &args)
	if err != nil {
		server.SendErrorf("unable to deserialize body: %s", err)
		return
	}

	// Run the flow using base wrapper script (which is 'run' command
	// inside container).
	cmd := exec.Command("run", args.ToArray()...)

	out, err := cmd.StdoutPipe()
	if err != nil {
		server.SendErrorf("unable to allocate stdout pipe: %s", err)
		return
	}

	go r.streamLogs(ctx, server, out)

	// Run command asynchrously using cmd.Run() method because it updates
	// cmd.ProcessState upon process completion, so we can detect when
	// the process ends.
	go func() {
		err = cmd.Run()
		if err != nil {
			server.SendErrorf("unable to start TagUI: %s", err)
			return
		}
	}()

	log.Debugw("Running TagUI flow", "flow", args.Flow)

	for {
		select {
		// Waiting for parent (request) context to end or process state to
		// change - whichever comes first.
		case <-ctx.Done():
			log.Debug("TagUI disconnected - scheduling clean-up...")
			go r.scheduleCleanup()
			return
		default:
			if cmd.ProcessState != nil {
				if cmd.ProcessState.Exited() {
					log.Debug("TagUI run finished")
					return
				}
			}
		}
	}
}
