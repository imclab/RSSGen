package errhandler

import (
	"os"
	"fmt"
	"runtime"
	"path/filepath"
)

// Handle an error for the calling function
func Handle(msg string, err error) {
	HandleDepth(msg, err, 2)
}

// Handle an error for any function at <depth> from the top of the call stack
func HandleDepth(msg string, err error, depth int) {
	// If the error is non-nil
	if err != nil {
		// Find out who called it and where from, skip the top <depth> calls
		pc, file, line, ok := runtime.Caller(depth)
		// Parse out the filename and calling function
		filename := filepath.Base(file)
		callingFunc := runtime.FuncForPC(pc)
		callingFuncName := callingFunc.Name()
		// If we could retrieve the information then print a message and exit with an error
		if ok {
			fmt.Printf("%s:%s:%d: %s %s\n", filename, callingFuncName, line, msg, err)
			os.Exit(1)
		}
	}
}
