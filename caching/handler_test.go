package caching

import (
	"testing"
	"time"
)

func TestOptions_BuildHeaderValues(t1 *testing.T) {
	tests := []struct {
		name   string
		fields Options
		want   string
	}{
		{
			name: "all",
			fields: Options{
				Access:          AccessPublic,
				MaxAge:          5 * time.Minute,
				SMaxAge:         10 * time.Minute,
				NoCache:         true,
				NoStore:         true,
				MustRevalidate:  true,
				ProxyRevalidate: true,
				MustUnderstand:  true,
				NoTransform:     true,
				Immutable:       true,
			},
			want: "public, max-age=300, s-max-age=600, no-cache, no-store, must-revalidate, proxy-revalidate, must-understand, no-transform, immutable",
		},
		{
			name: "partial",
			fields: Options{
				Access:         AccessPublic,
				MaxAge:         30 * time.Second,
				MustRevalidate: true,
			},
			want: "public, max-age=30, must-revalidate",
		},
		{
			name:   "empty",
			fields: Options{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Options{
				Access:          tt.fields.Access,
				MaxAge:          tt.fields.MaxAge,
				SMaxAge:         tt.fields.SMaxAge,
				NoCache:         tt.fields.NoCache,
				NoStore:         tt.fields.NoStore,
				MustRevalidate:  tt.fields.MustRevalidate,
				ProxyRevalidate: tt.fields.ProxyRevalidate,
				MustUnderstand:  tt.fields.MustUnderstand,
				NoTransform:     tt.fields.NoTransform,
				Immutable:       tt.fields.Immutable,
			}
			if got := t.BuildHeaderValues(); got != tt.want {
				t1.Errorf("BuildHeaderValues() failed\n"+
					"\twas:   %v\n"+
					"\twant:  %v", got, tt.want)
			}
		})
	}
}
