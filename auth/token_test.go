package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAllMethods(t *testing.T) {
	manager := NewTokenFactory([]byte("super secret key"), 5*time.Second)
	var token, encoded string

	t.Run("generate new token", func(t *testing.T) {
		token = manager.Generate()
		if token == "" {
			t.Errorf("got empty token")
		}
		if len(token) != 64 {
			t.Errorf("got token length %q, expected 64", len(token))
		}
	})

	t.Run("encode token", func(t *testing.T) {
		encoded = manager.Encode(token)
		if encoded == token {
			t.Errorf("encoded token same as input token")
		}
	})

	t.Run("decode token", func(t *testing.T) {
		decoded, err := manager.Decode(encoded)
		assertNoError(t, err)
		if decoded == encoded {
			t.Errorf("decoded token same as encoded input token")
		}
		if decoded != token {
			t.Errorf("got decoded token %q, expected %q", decoded, token)
		}
	})

	t.Run("set session token header", func(t *testing.T) {
		rec := httptest.NewRecorder()
		manager.WriteSessionHeader(rec, token)
		res := rec.Result()
		prefix := "Bearer "
		header := res.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			t.Errorf("got header %q, expected %q prefix", header, prefix)
		}
		dec, err := manager.Decode(strings.TrimPrefix(header, prefix))
		assertNoError(t, err)
		if dec != token {
			t.Errorf("got decoded token %q, expected %q", dec, token)
		}
	})

	t.Run("set session token cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		manager.WriteSessionCookie(rec, token)
		res := rec.Result()
		var cookie *http.Cookie
		for _, c := range res.Cookies() {
			if c.Name == "session" {
				cookie = c
				break
			}
		}
		if cookie.Value == "" {
			t.Errorf("got empty cookie")
		}
		dec, err := manager.Decode(cookie.Value)
		assertNoError(t, err)
		if dec != token {
			t.Errorf("got decoded token %q, expected %q", dec, token)
		}
		// TODO: get session from request
	})
}
