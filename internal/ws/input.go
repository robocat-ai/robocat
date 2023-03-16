package ws

import (
	"context"
	"errors"
	"os"
	"path"
)

type RobocatInput struct {
	runner *RobocatRunner
}

func NewRobocatInput(runner *RobocatRunner) *RobocatInput {
	return &RobocatInput{
		runner: runner,
	}
}

func (r *RobocatInput) Handle(
	ctx context.Context,
	message *Message,
) {
	file, err := ParseFileFromMessage(message)
	if err != nil {
		message.ReplyWithError(err)
		return
	}

	if len(file.Path) == 0 {
		message.ReplyWithError(errors.New("file path must not be empty"))
		return
	}

	inputBasePath, err := r.runner.GetFlowBasePath("input")
	if err != nil {
		message.ReplyWithError(err)
		return
	}

	absolutePath := path.Join(inputBasePath, file.Path)

	log.Debugw("Creating directory for input", "path", path.Dir(absolutePath))

	err = os.MkdirAll(path.Dir(absolutePath), 0777)
	if err != nil {
		message.ReplyWithError(err)
		return
	}

	log.Debugw("Writing input", "path", absolutePath)

	err = os.WriteFile(absolutePath, file.Payload, 0644)
	if err != nil {
		message.ReplyWithError(err)
		return
	}

	message.Reply("status", "ok")

	log.Debugw("Written input", "file", file.Path, "len", len(file.Payload))
}
