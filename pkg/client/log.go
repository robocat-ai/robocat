package robocat

import "errors"

type RobocatLog struct {
	cursor int
	lines  []string
}

// Append a new line to the log stream.
// This method is used internally by the RobocatFlow to append lines
// to the stream.
func (l *RobocatLog) append(line string) {
	l.lines = append(l.lines, line)
}

// Try to get next line from the log stream.
// If there are no new lines - returns an error.
func (l *RobocatLog) Next() (line string, err error) {
	if len(l.lines) > l.cursor {
		line = l.lines[l.cursor]
		l.cursor = l.cursor + 1
	} else {
		err = errors.New("no new log lines")
	}

	return
}
