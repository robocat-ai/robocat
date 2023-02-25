package ws

import (
	"context"
	"encoding/json"
	"os/exec"
	"path"
	"path/filepath"
)

type RobocatRunner struct {
	abortScheduledCleanupSignal chan bool
	cleanupScheduled            bool
	input                       *RobocatInput

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRobocatRunner() *RobocatRunner {
	runner := &RobocatRunner{
		abortScheduledCleanupSignal: make(chan bool),
		cleanupScheduled:            false,
	}

	runner.input = NewRobocatInput(runner)

	return runner
}

func (r *RobocatRunner) GetInput() *RobocatInput {
	return r.input
}

func (r *RobocatRunner) GetFlowBasePath(elem ...string) (string, error) {
	finalPath, err := filepath.Abs("flow")
	if err != nil {
		return finalPath, err
	}

	for _, el := range elem {
		finalPath = path.Join(finalPath, el)
	}

	return finalPath, nil
}

func (r *RobocatRunner) Handle(
	ctx context.Context,
	message *Message,
) {
	r.ctx, r.cancel = context.WithCancel(ctx)

	// In case of quick disconnect right after connection TagUI flow can
	// still be running, so we try to kill previously running TagUI instance
	// using scheduled clean-up. However, when the flow is run again we must
	// abort scheduled clean-up and instead clean-up manually.
	r.abortScheduledCleanup()
	r.cleanup()

	var args *RunnerArguments

	err := json.Unmarshal(message.Body, &args)
	if err != nil {
		message.ReplyWithErrorf("unable to deserialize body: %s", err)
		r.cancel()
		return
	}

	// Run the flow using base wrapper script (which is 'run' command
	// inside container).
	cmd := exec.Command("run", args.ToArray()...)

	out, err := cmd.StdoutPipe()
	if err != nil {
		message.ReplyWithErrorf("unable to allocate stdout pipe: %s", err)
		r.cancel()
		return
	}

	go r.watchLogs(r.ctx, message, out)
	go r.watchOutput(r.ctx, message)

	// Run command asynchrously using cmd.Run() method because it updates
	// cmd.ProcessState upon process completion, so we can detect when
	// the process ends.
	go func() {
		err = cmd.Run()
		if err != nil {
			message.ReplyWithErrorf("unable to start TagUI: %s", err)
			return
		}
	}()

	log.Debugw("Running TagUI flow", "flow", args.Flow)

	message.Reply("status", "ok")

loop:
	for {
		select {
		// Waiting for parent (request) context to end or process state to
		// change - whichever comes first.
		case <-ctx.Done():
			log.Debug("TagUI disconnected - scheduling clean-up...")
			go r.scheduleCleanup()
			break loop
		case <-r.ctx.Done():
			log.Debug("Received stop signal - stopping...")
			go r.scheduleCleanup()
			break loop
		default:
			if cmd.ProcessState != nil {
				if cmd.ProcessState.Exited() {
					log.Debug("TagUI run finished")
					break loop
				}
			}
		}
	}

	r.cancel()
}

func (r *RobocatRunner) Stop(
	ctx context.Context,
	message *Message,
) {
	if r.ctx == nil || r.ctx.Err() != nil {
		log.Debug("TagUI run is not running - cannot stop")
		message.ReplyWithErrorf("flow is not running - cannot stop")
		return
	}

	log.Debug("Sending stop signal...")

	r.cancel()
	message.Reply("status", "ok")
}
