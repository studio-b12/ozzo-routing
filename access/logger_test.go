// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package access

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/stretchr/testify/assert"
)

func TestCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	var customFunc = func(req *http.Request, rw *LogResponseWriter, elapsed float64, ttfb float64) {
		var logWriter = getLogger(&buf)
		clientIP := GetClientIP(req)
		requestLine := fmt.Sprintf("%s %s %s", req.Method, req.URL.String(), req.Proto)
		logWriter(`[%s] [%.3fms] %s %d %d %.3fms`, clientIP, elapsed, requestLine, rw.Status, rw.BytesWritten, ttfb)
	}
	h := CustomLogger(customFunc)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
	c := routing.NewContext(res, req, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET http://127.0.0.1/users")
}

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	h := Logger(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
	c := routing.NewContext(res, req, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET http://127.0.0.1/users")
}

func TestLogResponseWriter(t *testing.T) {
	res := httptest.NewRecorder()
	w := &LogResponseWriter{ResponseWriter: res}
	w.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, http.StatusBadRequest, w.Status)
	n, _ := w.Write([]byte("test"))
	assert.Equal(t, 4, n)
	assert.Equal(t, int64(4), w.BytesWritten)
	assert.Equal(t, "test", res.Body.String())
}

func TestGetClientIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users/", nil)
	req.Header.Set("X-Real-IP", "192.168.100.1")
	req.Header.Set("X-Forwarded-For", "192.168.100.2")
	req.RemoteAddr = "192.168.100.3"

	assert.Equal(t, "192.168.100.1", GetClientIP(req))
	req.Header.Del("X-Real-IP")
	assert.Equal(t, "192.168.100.2", GetClientIP(req))
	req.Header.Del("X-Forwarded-For")
	assert.Equal(t, "192.168.100.3", GetClientIP(req))

	req.RemoteAddr = "192.168.100.3:8080"
	assert.Equal(t, "192.168.100.3", GetClientIP(req))
}

func getLogger(buf *bytes.Buffer) LogFunc {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(buf, format, a...)
	}
}

func handler1(c *routing.Context) error {
	return errors.New("abc")
}

// Middlewares registered after the logger (e.g. for metrics) must see the time to first byte
// right after c.Next(), not only the logger itself after the whole request.
func TestTimeToFirstByteVisibleInsideLogger(t *testing.T) {
	const before, after = 30 * time.Millisecond, 150 * time.Millisecond

	var inner, logged time.Duration
	r := routing.New()
	r.Use(
		CustomLogger(func(req *http.Request, rw *LogResponseWriter, elapsedMs float64, ttfbMs float64) {
			logged = rw.TimeToFirstByte
		}),
		func(c *routing.Context) error {
			err := c.Next()
			inner = c.Response.(*LogResponseWriter).TimeToFirstByte
			return err
		},
	)
	r.To("GET", "/", func(c *routing.Context) error {
		time.Sleep(before)
		c.Response.WriteHeader(http.StatusOK)
		time.Sleep(after)
		_, err := c.Response.Write([]byte("body"))
		return err
	})
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if inner < before || inner >= before+after {
		t.Fatalf("time to first byte inside the logger = %v, want at least %v and clearly below %v", inner, before, before+after)
	}
	if logged != inner {
		t.Fatalf("logger saw %v, inner middleware saw %v, want the same value", logged, inner)
	}
}

// Without any write, the logger still reports the total duration as time to first byte.
func TestTimeToFirstByteWithoutWrite(t *testing.T) {
	const wait = 20 * time.Millisecond

	var logged time.Duration
	r := routing.New()
	r.Use(CustomLogger(func(req *http.Request, rw *LogResponseWriter, elapsedMs float64, ttfbMs float64) {
		logged = rw.TimeToFirstByte
	}))
	r.To("GET", "/", func(c *routing.Context) error {
		time.Sleep(wait)
		return nil
	})
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if logged < wait {
		t.Fatalf("time to first byte = %v, want at least %v", logged, wait)
	}
}

// A writer not created by CustomLogger has no start time and leaves TimeToFirstByte untouched.
func TestTimeToFirstByteWithoutLogger(t *testing.T) {
	w := &LogResponseWriter{ResponseWriter: httptest.NewRecorder()}
	w.WriteHeader(http.StatusOK)
	if w.FirstWrite.IsZero() {
		t.Fatal("FirstWrite not set")
	}
	if w.TimeToFirstByte != 0 {
		t.Fatalf("TimeToFirstByte = %v, want 0 without a start time", w.TimeToFirstByte)
	}
}
