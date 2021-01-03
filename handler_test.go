package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lmas/web/internal/assert"
	"github.com/pkg/errors"
)

func testHandler(t *testing.T, method, path string, f func(*Context) error) *Handler {
	t.Helper()
	h := NewHandler(nil)
	if f != nil {
		h.Register(method, path, f)
	}
	h.mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, ret interface{}) {
		t.Fatalf("Panic: %+v", ret)
	}
	return h
}

////////////////////////////////////////////////////////////////////////////////

func TestSimple(t *testing.T) {
	t.Run("simple get with no routes", func(t *testing.T) {
		h := testHandler(t, "", "", nil)
		resp := assert.DoRequest(t, h, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusNotFound)
	})
	t.Run("simple get", func(t *testing.T) {
		h := testHandler(t, "GET", "/hello", func(ctx *Context) error {
			fmt.Fprintf(ctx.W, "hello world")
			return nil
		})
		resp := assert.DoRequest(t, h, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world")
	})
	t.Run("get http error", func(t *testing.T) {
		h := testHandler(t, "GET", "/hello", func(ctx *Context) error {
			return ctx.Error(http.StatusNotImplemented, "test")
		})
		resp := assert.DoRequest(t, h, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusNotImplemented)
		assert.Body(t, resp, "Error: \"test\"\n")
	})
	t.Run("get unknown error", func(t *testing.T) {
		h := testHandler(t, "GET", "/hello", func(ctx *Context) error {
			return errors.New("test")
		})
		resp := assert.DoRequest(t, h, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusInternalServerError)
		assert.Body(t, resp, "Internal Server Error\n")
	})
	t.Run("panic in a handler", func(t *testing.T) {
		msg := "test"
		h := testHandler(t, "GET", "/hello", func(ctx *Context) error {
			panic(msg)
		})
		h.mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, ret interface{}) {
			s, ok := ret.(string)
			if !ok {
				t.Errorf("got panic value %q, expected %q", ret, msg)
			}
			http.Error(w, s, http.StatusInternalServerError)
		}
		resp := assert.DoRequest(t, h, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusInternalServerError)
		assert.Body(t, resp, msg+"\n")
	})
}

func TestRegisterPrefix(t *testing.T) {
	t.Run("register simple prefix", func(t *testing.T) {
		h := testHandler(t, "", "", nil)
		f := h.RegisterPrefix("/api")
		hello := func(msg string) func(*Context) error {
			return func(ctx *Context) error {
				fmt.Fprintf(ctx.W, msg)
				return nil
			}
		}
		f("GET", "/hello", hello("hello world"))
		f("GET", "/hello2", hello("hello world2"))
		f("GET", "/error", func(ctx *Context) error {
			return errors.New("test")
		})

		resp := assert.DoRequest(t, h, "GET", "/api/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world")

		resp = assert.DoRequest(t, h, "GET", "/api/hello2", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world2")

		resp = assert.DoRequest(t, h, "GET", "/api/error", nil, nil)
		assert.StatusCode(t, resp, http.StatusInternalServerError)
		assert.Body(t, resp, "Internal Server Error\n")
	})
	t.Run("register prefix with middleware", func(t *testing.T) {
		h := testHandler(t, "", "", nil)
		mw := func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("X-MSG", "hello")
				h.ServeHTTP(w, r)
			})
		}
		f := h.RegisterPrefix("/api", mw)
		hello := func(user string) func(*Context) error {
			return func(ctx *Context) error {
				msg := ctx.R.Header.Get("X-MSG")
				fmt.Fprintf(ctx.W, "%s %s", msg, user)
				return nil
			}
		}
		f("GET", "/hello", hello("world"))
		f("GET", "/hello2", hello("world2"))

		resp := assert.DoRequest(t, h, "GET", "/api/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world")

		resp = assert.DoRequest(t, h, "GET", "/api/hello2", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world2")
	})
}

////////////////////////////////////////////////////////////////////////////////

func newBenchmarkHandler(b *testing.B) *Handler {
	h := NewHandler(nil)
	h.mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, ret interface{}) {
		b.Fatalf("Panic: %+v", ret)
	}
	return h
}

func BenchmarkHandler(b *testing.B) {
	h := newBenchmarkHandler(b)
	h.Register("GET", "/hello", func(ctx *Context) error {
		fmt.Fprint(ctx.W, "hello world")
		return nil
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.ServeHTTP(w, r)
	}
}
