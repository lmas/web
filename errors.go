package web

import (
	"fmt"
	"runtime"
)

type stack struct {
	pc []uintptr
}

func newStack(n int) *stack {
	pc := make([]uintptr, 10)
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

type errStack struct {
	stack *stack
}

func newErrStack(n int) *errStack {
	return &errStack{newStack(n)}
}

func (e *errStack) Stack() string {
	return e.stack.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// ErrorPanic is a custom error that contains the error value passed to a panic(). Whenever a panic is raised in a
// Handler, the panic will be recovered using HandlePanic.
type ErrorPanic struct {
	*errStack
	panic interface{}
}

// NewErrorPanic creates a new ErrorPanic using an error value returned from a recovered panic.
func NewErrorPanic(ret interface{}) *ErrorPanic {
	return &ErrorPanic{newErrStack(3), ret}
}

func (e *ErrorPanic) Error() string {
	return fmt.Sprintf("%+v", e.panic)
}

// ErrorClient is a custom error which also contains a http status code. Whenever a Handler returns this error, the error
// message and status code will be sent back to the client unaltered.
type ErrorClient struct {
	*errStack
	status int
	msg    string
}

// NewErrorClient creates a new ErrorClient using a http status code and msg body, to be displayed to a client.
func NewErrorClient(status int, msg string) *ErrorClient {
	return &ErrorClient{newErrStack(3), status, msg}
}

func (e *ErrorClient) Error() string {
	return e.msg
}

// Status returns the HTTP status code for this error.
func (e *ErrorClient) Status() int {
	return e.status
}

// ErrorServer is a custom error for internal server errors. It will always return a "500 internal server error" to
// clients and a server log produced with a specified message.
type ErrorServer struct {
	*errStack
	msg string
}

// NewErrorServer creates a new ErrorServer using a msg, that will be logged.
func NewErrorServer(msg string) *ErrorServer {
	return &ErrorServer{newErrStack(3), msg}
}

func (e *ErrorServer) Error() string {
	return e.msg
}
