package ws

import (
	"os/exec"
	"time"
)

func (r *RobocatRunner) scheduleCleanup() {
	r.cleanupScheduled = true

	log.Debug("Clean-up scheduled")

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
