package ws

import (
	"context"
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
	fields, err := ParseDataFields(message)
	if err != nil {
		message.ReplyWithError(err)
		return
	}

	inputBasePath, err := r.runner.GetFlowBasePath("input")
	if err != nil {
		message.ReplyWithError(err)
		return
	}

	filePath := path.Join(inputBasePath, fields.Path)

	err = os.MkdirAll(path.Dir(filePath), 0755)
	if err != nil {
		message.ReplyWithError(err)
		return
	}

	err = os.WriteFile(filePath, fields.Payload, 0644)
	if err != nil {
		message.ReplyWithError(err)
		return
	}

	message.Reply("status", "ok")

	log.Debugw("Written input", "path", fields.Path)
}
