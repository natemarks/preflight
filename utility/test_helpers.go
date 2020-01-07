package utility

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ThisFunctionName returns the name of the current function relative to the package.
// If I called this function in github.com/natemarks/preflight/utility.TestThisFunctionName
// it would just return the last part:
// utility.TestThisFunctionName
func ThisFunctionName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	result := fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function)
	SplitResult := strings.Split(result, "/")
	return SplitResult[len(SplitResult)-1]
}

// CallerFunctionName returns the name of the function that called the current function.
// the returned string format is the same as that of ThisFunctionName()
func CallerFunctionName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	result := fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function)
	SplitResult := strings.Split(result, "/")
	return SplitResult[len(SplitResult)-1]
}

// wrap a function call to get the stdout  as string data
func CapOut(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	os.Stdout = w

	defer func() {
		os.Stdout = stdout
	}()
	f()
	err = w.Close()
	if err != nil {
		log.Error(fmt.Sprintf("Unable to close config file"))
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to copy buffer contents"))
	}

	return buf.String()

}
