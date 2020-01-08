package web

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

// HTTP Basic Auth
// For more information, see:
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication

// BasicAuthFunc is a function signature used for check if a username/password
// is valid during HTTP Basic Auth.
// It should return true if user/pass is valid.
type BasicAuthFunc func(string, string) bool

// DefaultBasicAuth returns a BasicAuthFunc that securely checks if the provided
// username/password matches the defaults.
func DefaultBasicAuth(defaultUser, defaultPass string) BasicAuthFunc {
	// Time taken by subtle.ConstantTimeCompare() depends on the length of
	// the byte slices, so to prevent timing based attacks we hash the strings
	// See https://stackoverflow.com/a/39591234
	du := sha256.Sum256([]byte(defaultUser))
	dp := sha256.Sum256([]byte(defaultPass))

	return func(user, pass string) bool {
		u := sha256.Sum256([]byte(user))
		p := sha256.Sum256([]byte(pass))

		// Sum256 returns [size]byte's, so quickly convert them to []byte's
		// with a simple slice[:] call
		if subtle.ConstantTimeCompare(du[:], u[:]) == 1 &&
			subtle.ConstantTimeCompare(dp[:], p[:]) == 1 {
			return true
		}
		return false
	}
}

func requestBasicAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (h *Handler) checkBasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !h.opt.BasicAuth(user, pass) {
			requestBasicAuth(w, r)
			return
		}

		// 's all good ya'll
		next.ServeHTTP(w, r)
	})
}
