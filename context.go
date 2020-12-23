package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Context is a convenience struct for easing the handling of a http request.
type Context struct {
	H *Handler
	W http.ResponseWriter
	R *http.Request
	P httprouter.Params
}

////////////////////////////////////////////////////////////////////////////////
// Utilize a sync.Pool to keep recently unused context objects for later reuse.

func (h *Handler) newContext() *Context {
	return &Context{
		H: h,
	}
}

func (h *Handler) getContext(w http.ResponseWriter, r *http.Request, p httprouter.Params) *Context {
	c := h.contextPool.Get().(*Context)
	c.W = w
	c.R = r
	c.P = p
	return c
}

func (h *Handler) putContext(c *Context) {
	h.contextPool.Put(c)
}

////////////////////////////////////////////////////////////////////////////////

// SetHeader is a shortcut to set a header value for a response.
func (c *Context) SetHeader(key, value string) {
	c.W.Header().Set(key, value)
}

// GetHeader is a shortcut to get a header value from a request.
func (c *Context) GetHeader(key string) string {
	return c.R.Header.Get(key)
}

////////////////////////////////////////////////////////////////////////////////

// GetParams is a shortcut to get URL params, first one given by key.
// See https://pkg.go.dev/github.com/julienschmidt/httprouter#Param for more info.
func (c *Context) GetParams(key string) string {
	return c.P.ByName(key)
}

////////////////////////////////////////////////////////////////////////////////

// Error returns a special http error, for which you can specify the http
// response status.
func (c *Context) Error(status int, msg string) error {
	return &httpError{status, msg}
}

// Empty let's you send a response code with empty body.
func (c *Context) Empty(status int) error {
	c.W.WriteHeader(status)
	return nil
}

// ErrInvalidRedirectCode is returned from Redirect when the redirection code
// is out of range (300 to 308).
var ErrInvalidRedirectCode = errors.New("invalid redirect code")

// Redirect sends a redirection response.
// It also makes sure the response code is within range.
func (c *Context) Redirect(status int, url string) error {
	if status < 300 || status > 308 {
		return ErrInvalidRedirectCode
	}
	c.SetHeader("Location", url)
	c.W.WriteHeader(status)
	return nil
}

// Bytes is a quick helper to send a response, with the contents of []byte{} as body.
// You should set a Content-Type header yourself.
func (c *Context) Bytes(status int, data []byte) error {
	c.W.WriteHeader(status)
	_, err := c.W.Write(data)
	return err
}

// String is a helper to send a simple string body, with a 'text/plain' Content-Type
// header.
func (c *Context) String(status int, data string) error {
	c.SetHeader("Content-Type", "text/plain; charset=UTF-8")
	return c.Bytes(status, []byte(data))
}

// HTML is a helper to send a simple HTML body, with a 'text/html' Content-Type
// header.
func (c *Context) HTML(status int, data string) error {
	c.SetHeader("Content-Type", "text/html; charset=UTF-8")
	return c.Bytes(status, []byte(data))
}

// Stream tries to stream the contents of an 'io.Reader'.
// No Content-Type is auto detected, so you should set it yourself.
// Warning: if http.Server timeouts are set too short, this write might time out.
func (c *Context) Stream(status int, r io.Reader) error {
	c.W.WriteHeader(status)
	_, err := io.Copy(c.W, r)
	return err
}

// File attempts to send a file (located at path).
// Content-Type is autodetected and errors will be handled with a 'http.Error'.
// See 'http.ServeFile' and 'http.ServeContent' for more info.
// Warning: if http.Server timeouts are set too short, this write might time out.
func (c *Context) File(status int, path string) error {
	http.ServeFile(c.W, c.R, path) // doesn't return any errors, handled with http.Error response
	return nil
}

// JSON is a helper for JSON encoding the data and sending it with a response status.
func (c *Context) JSON(status int, data interface{}) error {
	// Basicly copied from http.Error()
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.SetHeader("X-Content-Type-Options", "nosniff")
	c.W.WriteHeader(status)
	return json.NewEncoder(c.W).Encode(data)
}

// DecodeJSON is a helper for JSON decoding a request body.
func (c *Context) DecodeJSON(data interface{}) error {
	defer c.R.Body.Close()
	return json.NewDecoder(c.R.Body).Decode(data)
}
