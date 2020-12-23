package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lmas/web/internal/assert"
	"github.com/pkg/errors"
)

func TestWriteBytesAndStrings(t *testing.T) {
	t.Run("write bytes", func(t *testing.T) {
		msg := []byte("hello world\n")
		h := testHandler(t, "GET", "/hello", func(ctx *Context) error {
			ctx.Bytes(200, msg)
			return nil
		})
		resp := assert.DoRequest(t, h, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, string(msg))
	})
	t.Run("write string", func(t *testing.T) {
		msg := "hello world\n"
		h := testHandler(t, "GET", "/hello", func(ctx *Context) error {
			ctx.String(200, msg)
			return nil
		})
		resp := assert.DoRequest(t, h, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, msg)
	})
}

func TestJSON(t *testing.T) {
	msg := "hello world"
	t.Run("get json", func(t *testing.T) {
		h := testHandler(t, "GET", "/json", func(ctx *Context) error {
			return ctx.JSON(http.StatusOK, msg)
		})
		resp := assert.DoRequest(t, h, "GET", "/json", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, fmt.Sprintf("%q\n", msg))
	})
	t.Run("post json", func(t *testing.T) {
		h := testHandler(t, "POST", "/json", func(ctx *Context) error {
			var ret string
			err := ctx.DecodeJSON(&ret)
			if err != nil {
				return ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "decoding body").Error())
			}
			if ret != msg {
				return ctx.Error(http.StatusBadRequest, errors.New("body mismatch").Error())
			}
			return nil
		})
		resp := assert.DoRequest(t, h, "POST", "/json", nil, strings.NewReader(fmt.Sprintf("%q\n", msg)))
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "")
	})
}

////////////////////////////////////////////////////////////////////////////////

func BenchmarkContextError(b *testing.B) {
	// This benchmark is pretty useless but eeeh... might aswell keep it
	// as an example performance goal for the other benchmarks
	// (check the output results, they're pretty much maxed...)
	h := newBenchmarkHandler(b)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := h.getContext(w, r, nil)
	msg := "hello world"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Error(200, msg)
	}
}

func BenchmarkContextBytes(b *testing.B) {
	h := newBenchmarkHandler(b)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := h.getContext(w, r, nil)
	msg := []byte("hello world")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Bytes(200, msg)
	}
}

func BenchmarkContextString(b *testing.B) {
	h := newBenchmarkHandler(b)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := h.getContext(w, r, nil)
	msg := "hello world"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.String(200, msg)
	}
}

func BenchmarkContextJSON(b *testing.B) {
	h := newBenchmarkHandler(b)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := h.getContext(w, r, nil)
	msg := "hello world"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.JSON(200, msg)
	}
}
