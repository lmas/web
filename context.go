package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
	P httprouter.Params
}

////////////////////////////////////////////////////////////////////////////////

type httpError struct {
	status int
	msg    string
}

func (e *httpError) Error() string {
	return e.msg
}

func (e *httpError) String() string {
	return fmt.Sprintf("Error: %q", e.msg)
}

func (e *httpError) Status() int {
	return e.status
}

func (c Context) Error(status int, msg string) error {
	return &httpError{status, msg}
}

////////////////////////////////////////////////////////////////////////////////

//func (c Context) JsonError(status int, msg string) error {
//// Ignore any json errors on purpose here, since we're already erroring out
//_ = c.Json(status, "error: "+msg)
//}

func (c Context) Json(status int, data interface{}) error {
	// Basicly copied from http.Error()
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.SetHeader("X-Content-Type-Options", "nosniff")
	c.W.WriteHeader(status)
	return json.NewEncoder(c.W).Encode(data)
}

func (c Context) DecodeJson(data interface{}) error {
	defer c.R.Body.Close()
	return json.NewDecoder(c.R.Body).Decode(data)
}

////////////////////////////////////////////////////////////////////////////////

func (c Context) SetHeader(key, value string) {
	c.W.Header().Set(key, value)
}

func (c Context) GetHeader(key string) string {
	return c.R.Header.Get(key)
}
