package logging

import "io"

type LogWriter struct {
	io.Writer
}
