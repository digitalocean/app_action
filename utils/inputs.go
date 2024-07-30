package utils

import (
	"fmt"
	"strconv"

	gha "github.com/sethvargo/go-githubactions"
)

// InputAsString parses the input as a string and sets the target.
func InputAsString(a *gha.Action, input string, required bool, target *string) error {
	str := a.GetInput(input)
	if str == "" && required {
		return fmt.Errorf("input %q is required", input)
	}
	*target = str
	return nil
}

// InputAsBool parses the input as a boolean and sets the target.
func InputAsBool(a *gha.Action, input string, required bool, target *bool) error {
	str := a.GetInput(input)
	if str == "" {
		if required {
			return fmt.Errorf("input %q is required", input)
		}

		// If the input is not required, we default to false.
		*target = false
		return nil
	}
	val, err := strconv.ParseBool(str)
	if err != nil {
		return fmt.Errorf("failed to parse %q as a boolean: %v", input, err)
	}
	*target = val
	return nil
}
