package web

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/lmas/web/internal/assert"
	"github.com/pkg/errors"
)

func TestJSON(t *testing.T) {
	msg := "hello world"
	t.Run("get json", func(t *testing.T) {
		h := testHandler(t, "GET", "/json", func(ctx Context) error {
			return ctx.JSON(http.StatusOK, msg)
		})
		resp := assert.DoRequest(t, h, "GET", "/json", nil, nil)
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, fmt.Sprintf("%q\n", msg))
	})
	t.Run("post json", func(t *testing.T) {
		h := testHandler(t, "POST", "/json", func(ctx Context) error {
			var ret string
			err := ctx.DecodeJSON(&ret)
			if err != nil {
				return ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "decoding body").Error())
			}
			if ret != msg {
				return ctx.Error(http.StatusBadRequest, errors.New("body mismatch").Error())
			}
			return nil
		})
		resp := assert.DoRequest(t, h, "POST", "/json", nil, strings.NewReader(fmt.Sprintf("%q\n", msg)))
		assert.StatusCode(t, resp, http.StatusOK)
		assert.Body(t, resp, "")
	})
}
