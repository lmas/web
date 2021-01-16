package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lmas/web/internal/assert"
)

func TestMiddleware(t *testing.T) {
	msg, sender := "hello", "world"
	mw1 := func(next Handler) Handler {
		return Handler(func(c *Context) error {
			c.R.Header.Set("X-MSG", msg)
			return next(c)
		})
	}
	mw2 := func(next Handler) Handler {
		return Handler(func(c *Context) error {
			if c.GetHeader("X-MSG") != msg {
				panic("didn't find header")
			}
			return next(c)
		})
	}

	t.Run("register single middleware", func(t *testing.T) {
		f := func(c *Context) error {
			return c.String(200, c.R.Header.Get("X-MSG"))
		}
		m := testMux(t, "", "", nil)
		m.Register("GET", "/hello", f, mw1)
		resp := assert.DoRequest(t, m, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, msg)
	})
	t.Run("register many middleware", func(t *testing.T) {
		f := func(c *Context) error {
			return c.String(200, fmt.Sprintf("%s from %s", c.R.Header.Get("X-MSG"), c.P.ByName("sender")))
		}
		m := testMux(t, "", "", nil)
		m.Register("GET", "/hello/:sender", f, mw1, mw2)
		resp := assert.DoRequest(t, m, "GET", "/hello/"+sender, nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, msg+" from "+sender)
	})
	t.Run("error in handler", func(t *testing.T) {
		f := func(c *Context) error {
			return fmt.Errorf("err")
		}
		m := testMux(t, "", "", nil)
		m.Register("GET", "/hello/:sender", f, mw1, mw2)
		resp := assert.DoRequest(t, m, "GET", "/hello/"+sender, nil, nil)
		assert.StatusCode(t, resp, http.StatusInternalServerError)
		assert.Body(t, resp, "Internal Server Error\n")
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func BenchmarkMiddleware(b *testing.B) {
	m := newBenchmarkMux(b)
	mw1 := func(next Handler) Handler {
		return Handler(func(c *Context) error {
			return next(c)
		})
	}
	m.Register("GET", "/hello", func(c *Context) error {
		return nil
	}, mw1)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ServeHTTP(w, r)
	}
}
