package robocat

import (
	"github.com/robocat-ai/robocat/internal/ws"
)

type RobocatOutputStream struct {
	RobocatStream[*ws.RobocatFile]
}
