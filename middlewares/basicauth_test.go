package middlewares

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/lmas/web/internal/assert"
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
	basic := BasicAuth(user, pass)
	wrapped := basic(basicHandler)

	t.Run("simple valid login", func(t *testing.T) {
		resp := doRequest(t, wrapped, "GET", "/", headers, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "ok")
	})
	t.Run("missing auth header", func(t *testing.T) {
		resp := doRequest(t, wrapped, "GET", "/", nil, nil)
		assert.StatusCode(t, resp, http.StatusUnauthorized)
		assert.Header(t, resp, "WWW-Authenticate", `Basic realm="Restricted"`)
	})
	t.Run("wrong auth header", func(t *testing.T) {
		badHeaders := http.Header{
			"Authorization": []string{"Basic " + basicAuth("wrong", "wrong")},
		}
		resp := doRequest(t, wrapped, "GET", "/", badHeaders, nil)
		assert.StatusCode(t, resp, http.StatusUnauthorized)
		assert.Header(t, resp, "WWW-Authenticate", `Basic realm="Restricted"`)
	})
}
