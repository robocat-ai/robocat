package robocat

import (
	"github.com/robocat-ai/robocat/internal/ws"
)

type RobocatFileStream struct {
	RobocatStream[*ws.RobocatFile]
}
