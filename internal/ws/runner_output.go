package ws

import (
	"context"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
)

func (r *RobocatRunner) watchOutput(
	ctx context.Context,
	message *Message,
) {
	outputBasePath, err := r.GetFlowBasePath("output")
	if err != nil {
		log.Fatal(err)
	}

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

				message.Reply("output", RobocatDataFields{
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
