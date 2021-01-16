package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lmas/web/internal/assert"
)

func testMux(t *testing.T, method, path string, f func(*Context) error) *Mux {
	t.Helper()
	m := NewMux(nil)
	if f != nil {
		m.Register(method, path, f)
	}
	m.mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, ret interface{}) {
		t.Fatalf("Panic: %+v", ret)
	}
	return m
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func TestSimple(t *testing.T) {
	t.Run("simple get with no routes", func(t *testing.T) {
		m := testMux(t, "", "", nil)
		resp := assert.DoRequest(t, m, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusNotFound)
	})
	t.Run("simple get", func(t *testing.T) {
		m := testMux(t, "GET", "/hello", func(c *Context) error {
			return c.String(200, "hello world")
		})
		resp := assert.DoRequest(t, m, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world")
	})
	t.Run("get http error", func(t *testing.T) {
		m := testMux(t, "GET", "/hello", func(c *Context) error {
			return c.ErrorClient(http.StatusNotImplemented, "test")
		})
		resp := assert.DoRequest(t, m, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusNotImplemented)
		assert.Body(t, resp, "test\n")
	})
	t.Run("get unknown error", func(t *testing.T) {
		m := testMux(t, "GET", "/hello", func(c *Context) error {
			return fmt.Errorf("err")
		})
		resp := assert.DoRequest(t, m, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusInternalServerError)
		assert.Body(t, resp, "Internal Server Error\n")
	})
	t.Run("panic in a handler", func(t *testing.T) {
		msg := "test"
		m := testMux(t, "GET", "/hello", func(c *Context) error {
			panic(msg)
		})
		m.mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, ret interface{}) {
			s, ok := ret.(string)
			if !ok {
				t.Errorf("got panic value %q, expected %q", ret, msg)
			}
			http.Error(w, s, http.StatusInternalServerError)
		}
		resp := assert.DoRequest(t, m, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusInternalServerError)
		assert.Body(t, resp, msg+"\n")
	})
}

func TestRegisterPrefix(t *testing.T) {
	t.Run("register simple prefix", func(t *testing.T) {
		m := testMux(t, "", "", nil)
		f := m.RegisterPrefix("/api")
		hello := func(msg string) func(*Context) error {
			return func(c *Context) error {
				return c.String(200, msg)
			}
		}
		f("GET", "/hello", hello("hello world"))
		f("GET", "/hello2", hello("hello world2"))
		f("GET", "/error", func(c *Context) error {
			return fmt.Errorf("err")
		})

		resp := assert.DoRequest(t, m, "GET", "/api/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world")

		resp = assert.DoRequest(t, m, "GET", "/api/hello2", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world2")

		resp = assert.DoRequest(t, m, "GET", "/api/error", nil, nil)
		assert.StatusCode(t, resp, http.StatusInternalServerError)
		assert.Body(t, resp, "Internal Server Error\n")
	})
	t.Run("register prefix with middleware", func(t *testing.T) {
		m := testMux(t, "", "", nil)
		mw := func(next Handler) Handler {
			return Handler(func(c *Context) error {
				c.R.Header.Set("X-MSG", "hello")
				return next(c)
			})
		}
		f := m.RegisterPrefix("/api", mw)
		hello := func(user string) func(*Context) error {
			return func(c *Context) error {
				return c.String(200, fmt.Sprintf("%s %s", c.R.Header.Get("X-MSG"), user))
			}
		}
		f("GET", "/hello", hello("world"))
		f("GET", "/hello2", hello("world2"))

		resp := assert.DoRequest(t, m, "GET", "/api/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world")

		resp = assert.DoRequest(t, m, "GET", "/api/hello2", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "hello world2")
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func newBenchmarkMux(b *testing.B) *Mux {
	m := NewMux(nil)
	m.mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, ret interface{}) {
		b.Fatalf("Panic: %+v", ret)
	}
	return m
}

func BenchmarkMux(b *testing.B) {
	m := newBenchmarkMux(b)
	m.Register("GET", "/hello", func(c *Context) error {
		return nil
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ServeHTTP(w, r)
	}
}
