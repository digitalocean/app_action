package parser

import (
	"encoding/json"
	"log"

	"github.com/ParamPatel207/app_action/internal/parser_struct"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

//ParseJsonInput parses updated json file to yaml
func ParseAppSpecToYaml(appSpec *godo.AppSpec) ([]byte, error) {
	newYaml, err := yaml.Marshal(appSpec)
	if err != nil {
		return nil, errors.Wrap(err, "Error in building yaml")
	}
	return newYaml, nil
}

// ParseDeploymentSpec parses deployment array and retrieves appSpec of recent deployment
func ParseDeploymentSpec(apps []byte) ([]godo.Deployment, error) {
	var app []godo.Deployment
	err := json.Unmarshal(apps, &app)
	if err != nil {
		log.Fatal("Error in retrieving app spec: ", err)
	}
	return app, nil
}

// ParseAppSpec parses appSpec and returns array of apps
func ParseAppSpec(apps []byte) ([]godo.App, error) {
	var arr []godo.App
	err := json.Unmarshal(apps, &arr)
	if err != nil {
		return nil, errors.Wrap(err, "error in parsing data for AppId")
	}
	return arr, nil
}

// parseJsonInput takes the array of json object as input and unique name of users app as appName
//it parses the input and returns UpdatedRepo of the input
func ParseJsonInput(input string) ([]parser_struct.UpdatedRepo, error) {
	//takes care of empty json Deployment (use case where we redeploy using same app spec)
	var allRepos []parser_struct.UpdatedRepo
	err := json.Unmarshal([]byte(input), &allRepos)
	if err != nil {
		return nil, errors.Wrap(err, "error in parsing json data from file")
	}
	return allRepos, nil
}
