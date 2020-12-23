package assert

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// DoRequest records a client request for debug/test use
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

// Error asserts an error
func Error(t *testing.T, got error, want error) {
	t.Helper()
	if got != want {
		t.Errorf("got error %q, want %q", got, want)
	}
}

// StatusCode asserts status code from a response
func StatusCode(t *testing.T, got *http.Response, want int) {
	t.Helper()
	if got.StatusCode != want {
		t.Errorf("got status code %d, wanted %d", got.StatusCode, want)
	}
}

// Header asserts header from a response
func Header(t *testing.T, resp *http.Response, key, val string) {
	t.Helper()
	v := resp.Header.Get(key)
	if v != val {
		t.Errorf("got value %q for header %q, wanted value %q", v, key, val)
	}
}

// BodyEmpty asserts response body is empty
func BodyEmpty(t *testing.T, got *http.Response) {
	t.Helper()
	b, err := ioutil.ReadAll(got.Body)
	got.Body.Close()
	if err != nil {
		t.Errorf("got error %q", err)
	}
	if len(b) != 0 {
		t.Fatalf("got non-empty body")
	}
}

// Body asserts body from a response
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
