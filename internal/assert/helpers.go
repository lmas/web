package assert

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func DoRequest(t *testing.T, handler http.Handler, method, path string, headers http.Header, body io.Reader) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	if headers != nil {
		req.Header = headers
	}
	handler.ServeHTTP(rec, req)
	return rec.Result()
}

func Error(t *testing.T, got error, want error) {
	t.Helper()
	if got != want {
		t.Errorf("got error %q, want %q", got, want)
	}
}

func StatusCode(t *testing.T, got *http.Response, want int) {
	t.Helper()
	if got.StatusCode != want {
		t.Errorf("got status code %d, wanted %d", got.StatusCode, want)
	}
}

func Header(t *testing.T, resp *http.Response, key, val string) {
	t.Helper()
	v := resp.Header.Get(key)
	if v != val {
		t.Errorf("got value %q for header %q, wanted value %q", v, key, val)
	}
}

func Body(t *testing.T, got *http.Response, want string) {
	t.Helper()
	b, err := ioutil.ReadAll(got.Body)
	got.Body.Close()
	if err != nil {
		t.Errorf("got error %q", err)
	}
	if want != "" && len(b) < 1 {
		t.Fatalf("got empty body")
	}
	if string(b) != want {
		t.Errorf("got body %q, wanted %q", b, want)
	}
}
