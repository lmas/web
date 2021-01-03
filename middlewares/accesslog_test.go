package middlewares

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/lmas/web/internal/assert"
)

func TestAccessLog(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})
	var buf bytes.Buffer
	mw := AccessLog(log.New(&buf, "", 0))
	wrapped := mw(h)

	resp := assert.DoRequest(t, wrapped, "GET", "/", nil, nil)
	assert.StatusCode(t, resp, http.StatusOK)
	if len(buf.String()) < 1 {
		t.Errorf("didn't get a log line")
	}
}
