package main

import (
	"github.com/digitalocean/app_action/utils"
	gha "github.com/sethvargo/go-githubactions"
)

// inputs are the inputs for the action.
type inputs struct {
	token          string
	appName        string
	appID          string
	fromPRPreview  bool
	ignoreNotFound bool
}

// getInputs gets the inputs for the action.
func getInputs(a *gha.Action) (inputs, error) {
	var in inputs
	for _, err := range []error{
		utils.InputAsString(a, "token", true, &in.token),
		utils.InputAsString(a, "app_name", false, &in.appName),
		utils.InputAsString(a, "app_id", false, &in.appID),
		utils.InputAsBool(a, "from_pr_preview", false, &in.fromPRPreview),
		utils.InputAsBool(a, "ignore_not_found", false, &in.ignoreNotFound),
	} {
		if err != nil {
			return in, err
		}
	}
	return in, nil
}
