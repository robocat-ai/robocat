package robocat

import "errors"

type RobocatLog struct {
	cursor int
	lines  []string
}

func (l *RobocatLog) Append(line string) {
	l.lines = append(l.lines, line)
}

func (l *RobocatLog) Next() (line string, err error) {
	if len(l.lines) > l.cursor {
		line = l.lines[l.cursor]
		l.cursor = l.cursor + 1
	} else {
		err = errors.New("no new log lines")
	}

	return
}
