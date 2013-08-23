package gawker

import (
    "fmt"
    "os"
    "runtime"
    "strings"
)

var debug bool

// Debug function, if the debug flag is set, then display. Do nothing otherwise
// If Docker is in damon mode, also send the debug info on the socket
func Debugf(format string, a ...interface{}) {
    if debug {
        // Retrieve the stack infos
        _, file, line, ok := runtime.Caller(1)
        if !ok {
            file = "<unknown>"
            line = -1
        } else {
            file = file[strings.LastIndex(file, "/")+1:]
        }

        fmt.Fprintf(os.Stderr, fmt.Sprintf("[debug] %s:%d %s\n", file, line, format), a...)
    }
}
