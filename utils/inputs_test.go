package utils

import (
	"testing"

	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/require"
)

func TestInputAsString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		required bool
		expected string
		err      bool
	}{{
		name:     "success",
		input:    "input",
		required: true,
		expected: "value",
	}, {
		name:     "required",
		input:    "empty",
		required: true,
		err:      true,
	}, {
		name:     "optional",
		input:    "empty",
		required: false,
		expected: "",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := gha.New(gha.WithGetenv(func(k string) string {
				switch k {
				case "INPUT_INPUT":
					return "value"
				case "INPUT_EMPTY":
					return ""
				default:
					return "unexpected"
				}
			}))
			var target string
			err := InputAsString(a, test.input, test.required, &target)
			if !test.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Equal(t, test.expected, target)
		})
	}
}

func TestInputAsBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		required bool
		expected bool
		err      bool
	}{{
		name:     "success",
		input:    "input",
		required: true,
		expected: true,
	}, {
		name:     "required",
		input:    "empty",
		required: true,
		err:      true,
	}, {
		name:     "optional",
		input:    "empty",
		required: false,
		expected: false,
	}, {
		name:     "invalid",
		input:    "invalid",
		required: true,
		err:      true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := gha.New(gha.WithGetenv(func(k string) string {
				switch k {
				case "INPUT_INPUT":
					return "true"
				case "INPUT_EMPTY":
					return ""
				case "INPUT_INVALID":
					return "invalid"
				default:
					return "unexpected"
				}
			}))
			var target bool
			err := InputAsBool(a, test.input, test.required, &target)
			if !test.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Equal(t, test.expected, target)
		})
	}
}
