package web

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func TestBasicAuth(t *testing.T) {
	h := testHandler(t, "", "", nil)
	user, pass := "admin", "adminpass"
	h.opt.BasicAuth = DefaultBasicAuth(user, pass)
	// Notice we don't add any middlewares here, but still have the
	// checkBasicAuth middleware attached since we have set opt.BasicAuth.
	h.Register("GET", "/hello", func(ctx Context) error {
		fmt.Fprint(ctx.W, "ok")
		return nil
	})
	headers := http.Header{
		"Authorization": []string{"Basic " + basicAuth(user, pass)},
	}

	t.Run("simple valid login", func(t *testing.T) {
		resp := doRequest(t, h, "GET", "/hello", headers, nil)
		assertStatusCode(t, resp, http.StatusOK)
		assertBody(t, resp, "ok")
	})
	t.Run("missing auth header", func(t *testing.T) {
		resp := doRequest(t, h, "GET", "/hello", nil, nil)
		assertStatusCode(t, resp, http.StatusUnauthorized)
		assertHeader(t, resp, "WWW-Authenticate", `Basic realm="Restricted"`)
	})
	t.Run("wrong auth header", func(t *testing.T) {
		badHeaders := http.Header{
			"Authorization": []string{"Basic " + basicAuth("wrong", "wrong")},
		}
		resp := doRequest(t, h, "GET", "/hello", badHeaders, nil)
		assertStatusCode(t, resp, http.StatusUnauthorized)
		assertHeader(t, resp, "WWW-Authenticate", `Basic realm="Restricted"`)
	})
}
