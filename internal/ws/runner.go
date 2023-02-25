package ws

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
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
	input                       *RobocatInput
}

func NewRobocatRunner() *RobocatRunner {
	input := NewRobocatInput()

	return &RobocatRunner{
		abortScheduledCleanupSignal: make(chan bool),
		cleanupScheduled:            false,
		input:                       input,
	}
}

func (r *RobocatRunner) GetInput() *RobocatInput {
	return r.input
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

func (r *RobocatRunner) watchLogs(
	ctx context.Context,
	server *Server,
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
			server.Send("log", scanner.Text())
		}
	}

	log.Debug("Stopped watching logs")
}

func (r *RobocatRunner) watchOutput(
	ctx context.Context,
	server *Server,
) {
	flowBasePath, err := filepath.Abs("flow")
	if err != nil {
		log.Fatal(err)
	}

	outputBasePath := path.Join(flowBasePath, "output")

	w := watcher.New()

	w.FilterOps(watcher.Create, watcher.Write)

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.IsDir() {
					continue
				}

				log.Debugw("Got output update", "path", event.Path)

				path, err := filepath.Rel(outputBasePath, event.Path)
				if err != nil {
					log.Warnw("Unable to form relative path", "error", err)
					continue
				}

				ext := filepath.Ext(path)
				if len(ext) == 0 {
					ext = ".txt"
				}

				mimeType := mime.TypeByExtension(ext)

				payload, err := os.ReadFile(event.Path)
				if err != nil {
					log.Warnw("Unable to read file", "error", err, "file", event.Path)
					continue
				}

				server.Send("output", RobocatDataFields{
					Path:     path,
					MimeType: mimeType,
					Payload:  payload,
				})
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(outputBasePath); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}()
	defer w.Close()

	log.Debug("Watching output")

	// Wait until context is cancelled.
	<-ctx.Done()

	log.Debug("Stopped watching output")
}

func (r *RobocatRunner) Handle(
	ctx context.Context,
	server *Server,
	message *Message,
) {
	runnerCtx, runnerCtxCancel := context.WithCancel(ctx)

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
		runnerCtxCancel()
		return
	}

	// Run the flow using base wrapper script (which is 'run' command
	// inside container).
	cmd := exec.Command("run", args.ToArray()...)

	out, err := cmd.StdoutPipe()
	if err != nil {
		server.SendErrorf("unable to allocate stdout pipe: %s", err)
		runnerCtxCancel()
		return
	}

	go r.watchLogs(runnerCtx, server, out)
	go r.watchOutput(runnerCtx, server)

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

	server.Send("status", "ok")

loop:
	for {
		select {
		// Waiting for parent (request) context to end or process state to
		// change - whichever comes first.
		case <-ctx.Done():
			log.Debug("TagUI disconnected - scheduling clean-up...")
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

	runnerCtxCancel()
}
