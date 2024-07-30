package main

import (
	"github.com/digitalocean/app_action/utils"
	gha "github.com/sethvargo/go-githubactions"
)

// inputs are the inputs for the action.
type inputs struct {
	token           string
	appSpecLocation string
	appName         string
	printBuildLogs  bool
	printDeployLogs bool
	deployPRPreview bool
}

// getInputs gets the inputs for the action.
func getInputs(a *gha.Action) (inputs, error) {
	var in inputs
	for _, err := range []error{
		utils.InputAsString(a, "token", true, &in.token),
		utils.InputAsString(a, "app_spec_location", false, &in.appSpecLocation),
		utils.InputAsString(a, "app_name", false, &in.appName),
		utils.InputAsBool(a, "print_build_logs", true, &in.printBuildLogs),
		utils.InputAsBool(a, "print_deploy_logs", true, &in.printDeployLogs),
		utils.InputAsBool(a, "deploy_pr_preview", true, &in.deployPRPreview),
	} {
		if err != nil {
			return in, err
		}
	}
	return in, nil
}
