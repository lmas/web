package web

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestJson(t *testing.T) {
	msg := "hello world"
	t.Run("get json", func(t *testing.T) {
		h := testHandler(t, "GET", "/json", func(ctx Context) error {
			return ctx.Json(http.StatusOK, msg)
		})
		resp := doRequest(t, h, "GET", "/json", nil, nil)
		assertStatusCode(t, resp, http.StatusOK)
		assertBody(t, resp, fmt.Sprintf("%q\n", msg))
	})
	t.Run("post json", func(t *testing.T) {
		h := testHandler(t, "POST", "/json", func(ctx Context) error {
			var ret string
			err := ctx.DecodeJson(&ret)
			if err != nil {
				return ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "decoding body").Error())
			}
			if ret != msg {
				return ctx.Error(http.StatusBadRequest, errors.New("body mismatch").Error())
			}
			return nil
		})
		resp := doRequest(t, h, "POST", "/json", nil, strings.NewReader(fmt.Sprintf("%q\n", msg)))
		assertStatusCode(t, resp, http.StatusOK)
		assertBody(t, resp, "")
	})
}
