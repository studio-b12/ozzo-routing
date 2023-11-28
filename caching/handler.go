package caching

import (
	"fmt"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"strings"
	"time"
)

type AccessType string

const (
	AccessPublic  = AccessType("public")
	AccessPrivate = AccessType("private")
)

// Options defines Cache Control directive values as defined in
// RFC 7234 Section 5.2.1.
//
// https://www.rfc-editor.org/rfc/rfc7234#section-5.2.1
//
// Additional Reference:
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#directives
type Options struct {
	Access          AccessType
	MaxAge          time.Duration
	SMaxAge         time.Duration
	NoCache         bool
	NoStore         bool
	MustRevalidate  bool
	ProxyRevalidate bool
	MustUnderstand  bool
	NoTransform     bool
	Immutable       bool
}

func (t Options) BuildHeaderValues() string {
	var sb strings.Builder

	if t.Access != "" {
		fmt.Fprintf(&sb, "%s, ", t.Access)
	}

	if t.MaxAge != 0 {
		fmt.Fprintf(&sb, "max-age=%d, ", int64(t.MaxAge.Round(time.Second).Seconds()))
	}

	if t.SMaxAge != 0 {
		fmt.Fprintf(&sb, "s-max-age=%d, ", int64(t.SMaxAge.Round(time.Second).Seconds()))
	}

	if t.NoCache {
		fmt.Fprint(&sb, "no-cache, ")
	}

	if t.NoStore {
		fmt.Fprint(&sb, "no-store, ")
	}

	if t.MustRevalidate {
		fmt.Fprint(&sb, "must-revalidate, ")
	}

	if t.ProxyRevalidate {
		fmt.Fprint(&sb, "proxy-revalidate, ")
	}

	if t.MustUnderstand {
		fmt.Fprint(&sb, "must-understand, ")
	}

	if t.NoTransform {
		fmt.Fprint(&sb, "no-transform, ")
	}

	if t.Immutable {
		fmt.Fprint(&sb, "immutable, ")
	}

	v := sb.String()
	if len(v) < 2 {
		return ""
	}

	return v[:len(v)-2]
}

// Handler returns a routing.Handler which sets the "Cache-Control" header
// value as defined in the given options to the response.
func Handler(options Options) routing.Handler {
	headerValue := options.BuildHeaderValues()
	return func(c *routing.Context) error {
		c.Response.Header().Set("Cache-Control", headerValue)
		return nil
	}
}

// Public returns a routing.Handler which sets the "Cache-Control" header
// value to "public,max-age=<maxAge>".
func Public(maxAge time.Duration) routing.Handler {
	return Handler(Options{
		Access: AccessPublic,
		MaxAge: maxAge,
	})
}

// Private returns a routing.Handler which sets the "Cache-Control" header
// value to "private,max-age=<maxAge>".
func Private(maxAge time.Duration) routing.Handler {
	return Handler(Options{
		Access: AccessPrivate,
		MaxAge: maxAge,
	})
}

// NoCache returns a routing.Handler which sets the "Cache-Control" header
// value to "no-cache".
func NoCache() routing.Handler {
	return Handler(Options{NoCache: true})
}

// NoStore returns a routing.Handler which sets the "Cache-Control" header
// value to "no-store".
func NoStore() routing.Handler {
	return Handler(Options{NoStore: true})
}
