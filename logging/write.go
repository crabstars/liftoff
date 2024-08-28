package logging

import "fmt"

func (w LogWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}
