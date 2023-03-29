package ws

import (
	"bytes"
	"context"
	"fmt"
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
		log.Fatalw(err.Error(), "ref", message.Ref)
	}

	err = os.MkdirAll(outputBasePath, 0755)
	if err != nil {
		log.Warnw(err.Error(), "ref", message.Ref)
		return err
	}

	w := watcher.New()

	w.FilterOps(watcher.Create, watcher.Write)

	if err := w.AddRecursive(outputBasePath); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := w.Start(time.Millisecond * 300); err != nil {
			log.Fatalw(err.Error(), "ref", message.Ref)
		}
	}()
	defer w.Close()

	log.Debugw(fmt.Sprintf("Watching directory recusively: %s", outputBasePath), "ref", message.Ref)

	for {
		select {
		case event := <-w.Event:
			if event.IsDir() {
				continue
			}

			log.Debugw("Got output update", "path", event.Path, "ref", message.Ref)

			path, err := filepath.Rel(outputBasePath, event.Path)
			if err != nil {
				log.Warnw("Unable to form relative path", "error", err, "ref", message.Ref)
				continue
			}

			ext := filepath.Ext(path)
			if len(ext) == 0 {
				ext = ".txt"
			}

			mimeType := mime.TypeByExtension(ext)

			payload, err := os.ReadFile(event.Path)
			if err != nil {
				log.Warnw("Unable to read file", "error", err, "file", event.Path, "ref", message.Ref)
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
				log.Debugw(fmt.Sprintf("Output directory was removed: %s", outputBasePath), "ref", message.Ref)
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
				log.Warnw(fmt.Sprintf("Got output watcher error: %v", err), "ref", message.Ref)
			}
		}
	}

	log.Debugw("Stopped watching output", "ref", message.Ref)
}
