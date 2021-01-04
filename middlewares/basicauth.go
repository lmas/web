package middlewares

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

// HTTP Basic Auth
// For more information, see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication

// checkFunc is a function signature used for check if a username/password is valid during HTTP Basic Auth.
// It should return true on success.
type checkFunc func(string, string) bool

// singleBasicAuth returns a checkFunc that securely checks if the provided username/password matches the defaults.
func singleBasicAuth(defaultUser, defaultPass string) checkFunc {
	// Time taken by subtle.ConstantTimeCompare() depends on the length of the byte slices, so to prevent timing
	// based attacks we hash the strings. See https://stackoverflow.com/a/39591234
	du := sha256.Sum256([]byte(defaultUser))
	dp := sha256.Sum256([]byte(defaultPass))

	return func(user, pass string) bool {
		u := sha256.Sum256([]byte(user))
		p := sha256.Sum256([]byte(pass))

		// Sum256 returns [size]byte's, so quickly convert them to []byte's with a simple slice[:] call
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

////////////////////////////////////////////////////////////////////////////////////////////////////

// BasicAuth is a middleware that checks requests for the HTTP Basic Auth header and securely validates it against the
// given user/password pair.
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	isValid := singleBasicAuth(username, password)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || !isValid(user, pass) {
				requestBasicAuth(w, r)
				return
			}

			// 's all good ya'll
			next.ServeHTTP(w, r)
		})
	}
}
