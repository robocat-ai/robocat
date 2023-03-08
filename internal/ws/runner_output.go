package ws

import (
	"bytes"
	"context"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
)

func (r *RobocatRunner) watchOutputPath(
	ctx context.Context,
	message *Message,
	path string,
) error {
	outputBasePath, err := r.GetFlowBasePath(path)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(outputBasePath, 0755)
	if err != nil {
		log.Warn(err)
		return err
	}

	w := watcher.New()

	w.FilterOps(watcher.Create, watcher.Write)

	if err := w.AddRecursive(outputBasePath); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}()
	defer w.Close()

	log.Debugf("Watching directory recusively: %s", outputBasePath)

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

			if ext == ".txt" {
				payload = bytes.TrimSpace(payload)
			}

			message.Reply("output", RobocatFile{
				Path:     path,
				MimeType: mimeType,
				Payload:  payload,
			})
		case err := <-w.Error:
			if err == watcher.ErrWatchedFileDeleted {
				log.Debugf("Output directory was removed: %s", outputBasePath)
				return nil
			} else {
				return err
			}
		case <-ctx.Done():
		case <-w.Closed:
			return nil
		}
	}
}

func (r *RobocatRunner) watchOutput(
	ctx context.Context,
	message *Message,
) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			err := r.watchOutputPath(ctx, message, "output")
			if err != nil {
				log.Warnf("Got output watcher error: %w", err)
			}
		}
	}

	log.Debug("Stopped watching output")
}
