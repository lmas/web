package web

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
)

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got error %q", err)
	}
}

func testHandler(t *testing.T, method, path string, f func(Context) error) *Handler {
	t.Helper()
	//var buflog bytes.Buffer
	//h, err := New(log.New(&buflog, "", 0))
	//h, err := New(nil)
	//assertNoError(t, err)
	h := New(nil)
	if f != nil {
		h.Register(method, path, f)
	}
	h.mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, ret interface{}) {
		t.Fatalf("Panic: %+v", ret)
	}
	return h
}

func doRequest(t *testing.T, handler http.Handler, method, path string, headers http.Header, body io.Reader) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	if headers != nil {
		req.Header = headers
	}
	handler.ServeHTTP(rec, req)
	return rec.Result()
}

func assertStatusCode(t *testing.T, got *http.Response, want int) {
	t.Helper()
	if got.StatusCode != want {
		t.Errorf("got status code %d, wanted %d", got.StatusCode, want)
	}
}

func assertBody(t *testing.T, got *http.Response, want string) {
	t.Helper()
	b, err := ioutil.ReadAll(got.Body)
	got.Body.Close()
	assertNoError(t, err)
	if want != "" && len(b) < 1 {
		t.Fatalf("got empty body")
	}
	if string(b) != want {
		t.Errorf("got body %q, wanted %q", b, want)
	}
}

func assertHeader(t *testing.T, resp *http.Response, key, val string) {
	t.Helper()
	v := resp.Header.Get(key)
	if v != val {
		t.Errorf("got value %q for header %q, wanted value %q", v, key, val)
	}
}

////////////////////////////////////////////////////////////////////////////////

func TestSimple(t *testing.T) {
	t.Run("simple get with no routes", func(t *testing.T) {
		h := testHandler(t, "", "", nil)
		resp := doRequest(t, h, "GET", "/hello", nil, nil)
		assertStatusCode(t, resp, http.StatusNotFound)
	})
	t.Run("simple get", func(t *testing.T) {
		h := testHandler(t, "GET", "/hello", func(ctx Context) error {
			fmt.Fprintf(ctx.W, "hello world")
			return nil
		})
		resp := doRequest(t, h, "GET", "/hello", nil, nil)
		assertStatusCode(t, resp, http.StatusOK)
		assertBody(t, resp, "hello world")
	})
	t.Run("get http error", func(t *testing.T) {
		h := testHandler(t, "GET", "/hello", func(ctx Context) error {
			return ctx.Error(http.StatusNotImplemented, "test")
		})
		resp := doRequest(t, h, "GET", "/hello", nil, nil)
		assertStatusCode(t, resp, http.StatusNotImplemented)
		assertBody(t, resp, "Error: \"test\"\n")
	})
	t.Run("get unknown error", func(t *testing.T) {
		h := testHandler(t, "GET", "/hello", func(ctx Context) error {
			return errors.New("test")
		})
		resp := doRequest(t, h, "GET", "/hello", nil, nil)
		assertStatusCode(t, resp, http.StatusInternalServerError)
		assertBody(t, resp, "Internal Server Error\n")
	})
	t.Run("panic in a handler", func(t *testing.T) {
		msg := "test"
		h := testHandler(t, "GET", "/hello", func(ctx Context) error {
			panic(msg)
		})
		h.mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, ret interface{}) {
			s, ok := ret.(string)
			if !ok {
				t.Errorf("got panic value %q, expected %q", ret, msg)
			}
			http.Error(w, s, http.StatusInternalServerError)
		}
		resp := doRequest(t, h, "GET", "/hello", nil, nil)
		assertStatusCode(t, resp, http.StatusInternalServerError)
		assertBody(t, resp, msg+"\n")
	})
}

func TestRegisterPrefix(t *testing.T) {
	t.Run("register simple prefix", func(t *testing.T) {
		h := testHandler(t, "", "", nil)
		f := h.RegisterPrefix("/api")
		hello := func(msg string) func(Context) error {
			return func(ctx Context) error {
				fmt.Fprintf(ctx.W, msg)
				return nil
			}
		}
		f("GET", "/hello", hello("hello world"))
		f("GET", "/hello2", hello("hello world2"))
		f("GET", "/error", func(ctx Context) error {
			return errors.New("test")
		})

		resp := doRequest(t, h, "GET", "/api/hello", nil, nil)
		assertStatusCode(t, resp, http.StatusOK)
		assertBody(t, resp, "hello world")

		resp = doRequest(t, h, "GET", "/api/hello2", nil, nil)
		assertStatusCode(t, resp, http.StatusOK)
		assertBody(t, resp, "hello world2")

		resp = doRequest(t, h, "GET", "/api/error", nil, nil)
		assertStatusCode(t, resp, http.StatusInternalServerError)
		assertBody(t, resp, "Internal Server Error\n")
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
		hello := func(user string) func(Context) error {
			return func(ctx Context) error {
				msg := ctx.R.Header.Get("X-MSG")
				fmt.Fprintf(ctx.W, "%s %s", msg, user)
				return nil
			}
		}
		f("GET", "/hello", hello("world"))
		f("GET", "/hello2", hello("world2"))

		resp := doRequest(t, h, "GET", "/api/hello", nil, nil)
		assertStatusCode(t, resp, http.StatusOK)
		assertBody(t, resp, "hello world")

		resp = doRequest(t, h, "GET", "/api/hello2", nil, nil)
		assertStatusCode(t, resp, http.StatusOK)
		assertBody(t, resp, "hello world2")
	})
}
