package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Context is a convenience struct for easing the handling of a http request.
type Context struct {
	M *Mux
	W http.ResponseWriter
	R *http.Request
	P httprouter.Params
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Utilize a sync.Pool to keep recently unused context objects for later reuse.

func (m *Mux) newContext() interface{} {
	return &Context{
		M: m,
	}
}

func (m *Mux) getContext(w http.ResponseWriter, r *http.Request, p httprouter.Params) *Context {
	c := m.contextPool.Get().(*Context)
	c.W = w
	c.R = r
	c.P = p
	return c
}

func (m *Mux) putContext(c *Context) {
	c.W = nil
	c.R = nil
	c.P = nil
	m.contextPool.Put(c)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// SetHeader is a shortcut to set a header value for a response.
func (c *Context) SetHeader(key, value string) {
	c.W.Header().Set(key, value)
}

// GetHeader is a shortcut to get a header value from a request.
func (c *Context) GetHeader(key string) string {
	return c.R.Header.Get(key)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// GetParams is a shortcut to get URL params, first one given by key.
// See https://pkg.go.dev/github.com/julienschmidt/httprouter#Param for more info.
// It's safe to call when *Context.P == nil
func (c *Context) GetParams(key string) string {
	return c.P.ByName(key)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Error returns a special http error, for which you can specify the http response status.
func (c *Context) Error(status int, msg string) error {
	return &ErrorHTTP{status, msg}
}

// NotFound returns the result from the '404 not found' handler set at setup.
func (c *Context) NotFound() error {
	return c.M.opt.HandleNotFound(c)
}

// Empty let's you send a response code with empty body.
func (c *Context) Empty(status int) error {
	c.W.WriteHeader(status)
	return nil
}

// ErrInvalidRedirectCode is returned from Redirect() when the redirection code is out of range (300 to 308).
var ErrInvalidRedirectCode = errors.New("invalid redirect code")

// Redirect sends a redirection response. It also makes sure the response code is within range.
func (c *Context) Redirect(status int, url string) error {
	if status < 300 || status > 308 {
		return ErrInvalidRedirectCode
	}
	c.SetHeader("Location", url)
	c.W.WriteHeader(status)
	return nil
}

// Bytes is a quick helper to send a response, with the contents of []byte{} as body. You should set a Content-Type header yourself.
func (c *Context) Bytes(status int, data []byte) error {
	c.W.WriteHeader(status)
	_, err := c.W.Write(data)
	return err
}

// String is a helper to send a simple string body, with a 'text/plain' Content-Type header.
func (c *Context) String(status int, data string) error {
	c.SetHeader("Content-Type", "text/plain; charset=UTF-8")
	return c.Bytes(status, []byte(data))
}

// HTML is a helper to send a simple HTML body, with a 'text/html' Content-Type header.
func (c *Context) HTML(status int, data string) error {
	c.SetHeader("Content-Type", "text/html; charset=UTF-8")
	return c.Bytes(status, []byte(data))
}

// Stream tries to stream the contents of an 'io.Reader'. No Content-Type is auto detected, so you should set it
// yourself.
// Warning: if http.Server timeouts are set too short, this write might time out.
func (c *Context) Stream(status int, r io.Reader) error {
	c.W.WriteHeader(status)
	_, err := io.Copy(c.W, r)
	return err
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Stolen from https://pkg.go.dev/net/http#example-FileServer-DotFileHiding
func hasDotPrefix(path string) bool {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

// File attemps to send the contents of a file, located at `fp` and opened using `fs`.
// Default behaviour:
// * File path will be cleaned and resolved
// * Dotfiles will be ignored (or not, optionally)
// * Directory listnings will be ignored (or optionally serve an index file instead)
// * Any paths ignored will serve a `NotFound()` response instead
// * Content-Type will be autodetected (see `http.ServeContent` for more info and other extra handling)
//
// Warning: if http.Server timeouts are set too short, this write might time out.
// TODO: replace http.FileSystem with io/fs.FS when go1.16 lands
func (c *Context) File(fs http.FileSystem, fp string) error {
	fp = filepath.Clean(fp)
	// TODO: add option to disable this check
	if hasDotPrefix(fp) {
		return c.NotFound()
	}

	f, err := fs.Open(fp)
	if err != nil {
		if !os.IsNotExist(err) {
			c.M.logError("File", err)
		}
		return c.NotFound()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		c.M.logError("File", err)
		return c.NotFound()
	}
	if fi.IsDir() {
		// TODO: add option to try serve an index file instead
		return c.NotFound()
	}

	http.ServeContent(c.W, c.R, fi.Name(), fi.ModTime(), f)
	return nil // ServeContent will handle any errors with a http.Error, so we do nothing else
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Mux) newTemplateBuff() interface{} {
	return &bytes.Buffer{}
}

func (m *Mux) getTemplateBuff() *bytes.Buffer {
	return m.templatePool.Get().(*bytes.Buffer)
}

func (m *Mux) putTemplateBuff(b *bytes.Buffer) {
	b.Reset()
	m.templatePool.Put(b)
}

var (
	// ErrNoSuchTemplate is returned from Render() when trying to render a missing/invalid template.
	ErrNoSuchTemplate = errors.New("invalid template name")
)

// Render tries to render a HTML template (using tmpl as key for the Options.Template map, from the handler).
// Optional data can be provided for the template.
func (c *Context) Render(status int, tmpl string, data interface{}) error {
	t, found := c.M.opt.Templates[tmpl]
	if !found {
		// TODO: might want to show "invalid template name: the_name.html" instead
		return ErrNoSuchTemplate
	}
	// If there's any errors in the template we'll catch them here using a bytes.Buffer and don't risk messing up
	// the output to the client (by writing directly to context.W too soon). Using a pool should speed things up
	// too (and play nicer with the GC etc. etc.).
	buff := c.M.getTemplateBuff()
	defer c.M.putTemplateBuff(buff)
	if err := t.Execute(buff, data); err != nil {
		return err
	}

	c.SetHeader("Content-Type", "text/html; charset=UTF-8")
	c.W.WriteHeader(status)
	_, err := buff.WriteTo(c.W)
	return err
}

////////////////////////////////////////////////////////////////////////////////////////////////////

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
