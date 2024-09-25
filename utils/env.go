package utils

import (
	"fmt"
	"os"
	"strings"
)

// appWideVariables are the environment variables that are shared across all components.
// See https://docs.digitalocean.com/products/app-platform/how-to/use-environment-variables/#app-wide-variables.
var appWideVariables = map[string]struct{}{
	"APP_DOMAIN": {},
	"APP_URL":    {},
	"APP_NAME":   {},
}

// ExpandEnvRetainingBindables expands the environment variables in s, but it
// keeps bindable variables intact.
// Since bindable variables look like env vars, notation-wise, we just don't
// expand them at all.
func ExpandEnvRetainingBindables(s string) string {
	return os.Expand(s, func(name string) string {
		value := os.Getenv(name)
		if value == "" {
			if _, ok := appWideVariables[name]; ok || looksLikeBindable(name) {
				// If the environment variable is not set, keep the respective
				// reference intact.
				return fmt.Sprintf("${%s}", name)
			}
		}
		return value
	})
}

// looksLikeBindable returns true if the key looks like a bindable variable.
// Environment variables can't usually contain dots, so if they do, we're
// fairly confident that it's a bindable variable.
func looksLikeBindable(key string) bool {
	return strings.Contains(key, ".")
}
