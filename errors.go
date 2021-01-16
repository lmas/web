package web

import (
	"fmt"
	"runtime"
)

// ErrorPanic is a custom error that contains the error value passed to a panic(). Whenever a panic is raised in a
// Handler, the panic will be recovered using HandlePanic.
type ErrorPanic struct {
	stack *stack
	panic interface{}
}

func (e *ErrorPanic) Error() string {
	return fmt.Sprintf("%+v", e.panic)
}

func (e *ErrorPanic) String() string {
	return fmt.Sprintf("Panic: %+v", e.panic)
}

// ErrorHTTP is a custom error which also contains a http status code. Whenever a Handler returns this error, the error
// message and status code will be sent back to the client unaltered.
type ErrorHTTP struct {
	stack  *stack
	status int
	msg    string
}

func (e *ErrorHTTP) Error() string {
	return e.msg
}

func (e *ErrorHTTP) String() string {
	return fmt.Sprintf("Error: %q", e.msg)
}

// Status returns the HTTP status code for this error.
func (e *ErrorHTTP) Status() int {
	return e.status
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type stack struct {
	pc []uintptr
}

func newStack(n int) *stack {
	pc := make([]uintptr, 3)
	l := runtime.Callers(n, pc)
	pc = pc[:l]
	return &stack{
		pc: pc,
	}
}

func (st *stack) String() string {
	s := "Stack:"
	if len(st.pc) < 1 {
		return s + "\nnil"
	}
	frames := runtime.CallersFrames(st.pc)
	for {
		f, more := frames.Next()
		s += fmt.Sprintf("\n%s\n\t%s:%d", f.Function, f.File, f.Line)
		if !more {
			break
		}
	}
	return s
}
