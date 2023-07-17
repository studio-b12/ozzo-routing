package fault

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	routing "github.com/studio-b12/ozzo-routing"
)

func TestPanicHandler(t *testing.T) {
	var buf bytes.Buffer
	h := PanicHandler(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := routing.NewContext(res, req, h, handler3, handler2)
	err := c.Next()
	if assert.NotNil(t, err) {
		assert.Equal(t, "xyz", err.Error())
	}
	assert.NotEqual(t, "", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req, h, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, "", buf.String())

	buf.Reset()
	h2 := ErrorHandler(getLogger(&buf))
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req, h2, h, handler3, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "xyz", res.Body.String())
	assert.Contains(t, buf.String(), "panic_test.go")
	assert.Contains(t, buf.String(), "xyz")
}
