package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

//func main() {
//m := NewTokenFactory([]byte("not so secret key"), 1*time.Second)

//token := m.Generate()
//fmt.Println("TOKEN:", len(token), token)

//encoded := m.Encode(token)
//fmt.Println("ENCODED:", len(encoded), encoded)

////time.Sleep(2 * time.Second)
////m.timeFunc = func() time.Time {
////return time.Now().Add(2 * time.Second)
////}

//decoded, err := m.Decode(encoded)
//if err != nil {
//panic(err)
//}

//fmt.Println("DECODED:", len(decoded), decoded)
//if decoded != token {
//panic("mismatched tokens")
//}
//}

////////////////////////////////////////////////////////////////////////////////

// Heavily inspired by https://godoc.org/github.com/gorilla/securecookie

type TokenFactory struct {
	signingKey []byte
	expires    time.Duration

	// Allows us to replace this func so we later can run tests on the
	// expire times
	timeFunc func() time.Time
}

func NewTokenFactory(key []byte, expires time.Duration) *TokenFactory {
	return &TokenFactory{
		signingKey: key,
		expires:    expires,
		timeFunc:   timeNow,
	}
}

func (t *TokenFactory) Generate() string {
	// 512bits of random data should be enough
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		// Any errors with the RNG is a sign of a busted OS, in which
		// case we're screwed anyho
		panic(errors.Wrap(err, "system failure for crypto/rand"))
	}

	h := sha256.New()
	_, err := h.Write(b)
	if err != nil {
		// Pretty sure sha256.Write() can never fail
		panic(errors.Wrap(err, "system failure for crypto/sha256"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (t *TokenFactory) Encode(token string) string {
	if len(token) != 64 {
		// Tokens from New() should always end up being 64 bytes long
		// hex encoded sha256 hashes. Force a panic if a dev tries to
		// use some other kind of token (don't want to make too large
		// payloads or we might bust a http cookie, with a max size of
		// 4096 bytes).
		panic(errors.New("token not 64 bytes long"))
	}

	timestamp := t.timeFunc().Unix()
	payload := []byte(fmt.Sprintf("%d.%s", timestamp, token))
	mac := generateHMAC(payload, t.signingKey)
	payload = append(payload, fmt.Sprintf(".%s", mac)...)
	encoded := encodePayload(payload)
	return encoded
}

func (t *TokenFactory) Decode(encoded string) (string, error) {
	// Can't assume the length is any specific anymore, so don't check it

	payload, err := decodePayload(encoded)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode payload")
	}

	parts := bytes.SplitN(payload, []byte("."), 3)
	if len(parts) != 3 {
		return "", errors.New("failed to split payload")
	}
	// removes the mac+separator from the end of the payload
	payload = payload[:len(payload)-len(parts[2])-1]

	if !verifyHMAC(payload, t.signingKey, parts[2]) {
		return "", errors.New("failed to validate payload: invalid hmac")
	}

	timestamp, err := strconv.ParseInt(string(parts[0]), 10, 64)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse timestamp")
	}
	expires := t.timeFunc().Add(-t.expires).Unix()
	if timestamp < expires {
		return "", errors.New("timestamp expired")
	}

	if len(parts[1]) != 64 {
		// Should never happen as the hmac should fail before we're
		// down here, but eeeh never hurts I guess
		return "", errors.New("token not 64 bytes long")
	}
	return string(parts[1]), nil
}

////////////////////////////////////////////////////////////////////////////////

func (t *TokenFactory) WriteSessionHeader(w http.ResponseWriter, token string) {
	enc := t.Encode(token)
	w.Header().Set("Authorization", "Bearer "+enc)
}

func (t *TokenFactory) WriteSessionCookie(w http.ResponseWriter, token string) {
	c := &http.Cookie{
		Name:     "session",
		Value:    t.Encode(token),
		MaxAge:   int(t.expires / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		//Secure:   true, // can only be used for https (not for dev mode)
	}
	http.SetCookie(w, c)
}

func (t *TokenFactory) GetSessionToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token, err := t.Decode(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return "", errors.Wrap(err, "failed to decode header")
		}
		return token, nil
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		return "", errors.Wrap(err, "failed to get cookie")
	}
	token, err := t.Decode(cookie.Value)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode cookie")
	}
	return token, nil
}

////////////////////////////////////////////////////////////////////////////////

func timeNow() time.Time {
	return time.Now().UTC()
}

func generateHMAC(payload, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(payload)
	return h.Sum(nil)
}

func verifyHMAC(payload, key, mac []byte) bool {
	expected := generateHMAC(payload, key)
	return hmac.Equal(expected, mac)
}

func encodePayload(b []byte) string {
	//return hex.EncodeToString(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodePayload(s string) ([]byte, error) {
	//return hex.DecodeString(s)
	return base64.RawURLEncoding.DecodeString(s)
}

//const (
// 8 + 56 = 64 bytes in total
//bytesTokenTime int = 8 // uint64 = 8 bytes long
//bytesTokenRand int = 56
//)

//func NewToken(ts time.Time) string {
//buf := make([]byte, bytesTokenTime+bytesTokenRand)
//t := uint64(ts.UnixNano())
//binary.BigEndian.PutUint64(buf[:bytesTokenTime], t)
//_, err := rand.Read(buf[bytesTokenTime:])
//if err != nil {
//panic(err)
//}
//// Will always output a 64 chars hex string
//h := sha256.New()
//h.Write(buf)
//return hex.EncodeToString(h.Sum(nil))
//}
