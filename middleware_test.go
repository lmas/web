package web

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/lmas/web/internal/assert"
	"github.com/pkg/errors"
)

func TestMiddleware(t *testing.T) {
	msg, sender := "hello", "world"
	mw1 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-MSG", msg)
			h.ServeHTTP(w, r)
		})
	}
	mw2 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-MSG") != msg {
				panic("didn't find header")
			}
			h.ServeHTTP(w, r)
		})
	}

	t.Run("register single middleware", func(t *testing.T) {
		f := func(ctx Context) error {
			fmt.Fprintf(ctx.W, ctx.R.Header.Get("X-MSG"))
			return nil
		}
		h := testHandler(t, "", "", nil)
		h.Register("GET", "/hello", f, mw1)
		resp := assert.DoRequest(t, h, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, msg)
	})
	t.Run("register many middleware", func(t *testing.T) {
		f := func(ctx Context) error {
			fmt.Fprintf(ctx.W, "%s from %s", ctx.R.Header.Get("X-MSG"), ctx.P.ByName("sender"))
			return nil
		}
		h := testHandler(t, "", "", nil)
		h.Register("GET", "/hello/:sender", f, mw1, mw2)
		resp := assert.DoRequest(t, h, "GET", "/hello/"+sender, nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, msg+" from "+sender)
	})
	t.Run("error in handler", func(t *testing.T) {
		f := func(ctx Context) error {
			return errors.New("test")
		}
		h := testHandler(t, "", "", nil)
		h.Register("GET", "/hello/:sender", f, mw1, mw2)
		resp := assert.DoRequest(t, h, "GET", "/hello/"+sender, nil, nil)
		assert.StatusCode(t, resp, http.StatusInternalServerError)
		assert.Body(t, resp, "Internal Server Error\n")
	})
}
