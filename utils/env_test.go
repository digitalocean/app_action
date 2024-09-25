package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandEnvRetainingBindables(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Setenv("APP_URL", "baz")

	tests := []struct {
		name string
		in   string
		out  string
	}{{
		name: "simple",
		in:   "hello $FOO",
		out:  "hello bar",
	}, {
		name: "bindable",
		in:   "hello ${FOO.bar}",
		out:  "hello ${FOO.bar}",
	}, {
		name: "global bindable, unset",
		in:   "hello ${APP_DOMAIN}",
		out:  "hello ${APP_DOMAIN}",
	}, {
		name: "global bindable, overridden in env",
		in:   "hello ${APP_URL}",
		out:  "hello baz",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ExpandEnvRetainingBindables(test.in)
			require.Equal(t, test.out, got)
		})
	}
}
