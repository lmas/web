package basicauth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/lmas/web/assert"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func TestBasicAuth(t *testing.T) {
	user, pass := "admin", "adminpass"
	headers := http.Header{
		"Authorization": []string{"Basic " + basicAuth(user, pass)},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})
	basic := New(user, pass)
	wrapped := basic(handler)

	t.Run("simple valid login", func(t *testing.T) {
		resp := assert.DoRequest(t, wrapped, "GET", "/hello", headers, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "ok")
	})
	t.Run("missing auth header", func(t *testing.T) {
		resp := assert.DoRequest(t, wrapped, "GET", "/hello", nil, nil)
		assert.StatusCode(t, resp, http.StatusUnauthorized)
		assert.Header(t, resp, "WWW-Authenticate", `Basic realm="Restricted"`)
	})
	t.Run("wrong auth header", func(t *testing.T) {
		badHeaders := http.Header{
			"Authorization": []string{"Basic " + basicAuth("wrong", "wrong")},
		}
		resp := assert.DoRequest(t, wrapped, "GET", "/hello", badHeaders, nil)
		assert.StatusCode(t, resp, http.StatusUnauthorized)
		assert.Header(t, resp, "WWW-Authenticate", `Basic realm="Restricted"`)
	})
}
