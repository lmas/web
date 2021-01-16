package web

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lmas/web/internal/assert"
)

func TestSimpleResponses(t *testing.T) {
	m := testMux(t, "", "", nil)
	req, _ := http.NewRequest("GET", "/", nil)
	testContext := func(f func(*Context) error) *http.Response {
		rec := httptest.NewRecorder()
		c := m.getContext(rec, req, nil)
		assert.Error(t, f(c), nil)
		return rec.Result()
	}

	t.Run("write empty body", func(t *testing.T) {
		resp := testContext(func(c *Context) error {
			return c.Empty(200)
		})
		assert.StatusCode(t, resp, http.StatusOK)
		assert.BodyEmpty(t, resp)
	})
	t.Run("write bytes", func(t *testing.T) {
		msg := []byte("hello world\n")
		resp := testContext(func(c *Context) error {
			return c.Bytes(200, msg)
		})
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, string(msg))
	})
	t.Run("write string", func(t *testing.T) {
		msg := "hello world\n"
		resp := testContext(func(c *Context) error {
			return c.String(200, msg)
		})
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Header(t, resp, "Content-Type", "text/plain; charset=utf-8")
		assert.Body(t, resp, msg)
	})
	t.Run("write html", func(t *testing.T) {
		msg := "<html>hello world</html>\n"
		resp := testContext(func(c *Context) error {
			return c.HTML(200, msg)
		})
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Header(t, resp, "Content-Type", "text/html; charset=utf-8")
		assert.Body(t, resp, msg)
	})
	t.Run("write template", func(t *testing.T) {
		m.opt.Templates = map[string]*template.Template{
			"test": template.Must(template.New("test").Parse("hello {{.Name}}")),
		}
		name := "world"
		resp := testContext(func(c *Context) error {
			return c.Render(200, "test", map[string]string{
				"Name": name,
			})
		})
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Header(t, resp, "Content-Type", "text/html; charset=utf-8")
		assert.Body(t, resp, "hello world")
	})
	t.Run("write json", func(t *testing.T) {
		msg := "hello world"
		resp := testContext(func(c *Context) error {
			return c.JSON(200, msg)
		})
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Header(t, resp, "Content-Type", "application/json; charset=utf-8")
		assert.Body(t, resp, fmt.Sprintf("%q\n", msg))
	})
	t.Run("write not found", func(t *testing.T) {
		resp := testContext(func(c *Context) error {
			return c.NotFound()
		})
		assert.StatusCode(t, resp, http.StatusNotFound)
		assert.Header(t, resp, "Content-Type", "text/plain; charset=utf-8")
		assert.Body(t, resp, "404 page not found\n")
	})

	t.Run("get params", func(t *testing.T) {
		resp := testContext(func(c *Context) error {
			_ = c.GetParams("test")
			return nil
		})
		assert.StatusCode(t, resp, http.StatusOK)
		assert.BodyEmpty(t, resp)
	})
}

func TestDecodeJSON(t *testing.T) {
	m := testMux(t, "", "", nil)
	msg := "hello world"
	req, _ := http.NewRequest("GET", "/", strings.NewReader(fmt.Sprintf("%q\n", msg)))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()
	c := m.getContext(rec, req, nil)

	var s string
	if err := c.DecodeJSON(&s); err != nil {
		t.Fatalf("error decoding json body: %s", err)
	}
	if s != msg {
		t.Errorf("got json body %q, wanted %q", s, msg)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func BenchmarkContextError(b *testing.B) {
	// This benchmark is pretty useless but eeeh... might aswell keep it as an example performance goal for the
	// other benchmarks (check the output results, they're pretty much maxed...)
	m := newBenchmarkMux(b)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := m.getContext(w, r, nil)
	msg := "hello world"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Error(200, msg)
	}
}

func BenchmarkContextBytes(b *testing.B) {
	m := newBenchmarkMux(b)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := m.getContext(w, r, nil)
	msg := []byte("hello world")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Bytes(200, msg)
	}
}

func BenchmarkContextString(b *testing.B) {
	m := newBenchmarkMux(b)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := m.getContext(w, r, nil)
	msg := "hello world"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.String(200, msg)
	}
}

func BenchmarkContextRender(b *testing.B) {
	m := newBenchmarkMux(b)
	m.opt.Templates = map[string]*template.Template{
		"test": template.Must(template.New("test").Parse("hello world")),
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := m.getContext(w, r, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Render(200, "test", nil)
	}
}

func BenchmarkContextJSON(b *testing.B) {
	m := newBenchmarkMux(b)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	c := m.getContext(w, r, nil)
	msg := "hello world"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.JSON(200, msg)
	}
}
