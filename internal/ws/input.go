package ws

import "context"

type RobocatInput struct {
	//
}

func NewRobocatInput() *RobocatInput {
	return &RobocatInput{
		//
	}
}

func (r *RobocatInput) Handle(
	ctx context.Context,
	server *Server,
	message *Message,
) {
	//
}
