package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	"gopkg.in/yaml.v2"
)

func TestGetAllRepo(t *testing.T) {

	temp := `[ {
		"name": "frontend",
		"repository": "registry.digitalocean.com/<my-registry>/<my-image>",
		"tag": "latest"
	  }]`
	allRepos, err := getAllRepo(temp, "_")
	if err != nil {
		t.Errorf("Error in parsing input json data")
	}
	if allRepos[0].Name != "frontend" ||
		allRepos[0].Repository != "registry.digitalocean.com/<my-registry>/<my-image>" ||
		allRepos[0].Tag != "latest" {
		t.Errorf("Error in unmarshal")
	}
	//testing individual deployment for get all repo
	_, err = getAllRepo("", os.Getenv("TEST_APP_NAME"))
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestCheckForGitAndDockerHub(t *testing.T) {
	test_input, err := ioutil.ReadFile("sample-golang.yaml")
	if err != nil {
		t.Errorf("error in reading test file")
	}
	var app godo.AppSpec
	err = yaml.Unmarshal(test_input, &app)
	if err != nil {
		t.Errorf("Error in unmarshalling test yaml")
	}
	if app.Services[0].Name == "web" && app.Services[0].Git.RepoCloneURL == "https://github.com/snormore/sample-golang.git" {
		t.Errorf("Error in parsing test data")
	}
	temp := `[ {
		"name": "web",
		"repository": "registry.digitalocean.com/<my-registry>/<my-image>",
		"tag": "latest"
	  }]`
	allRepos, err := getAllRepo(temp, "_")
	if err != nil {
		t.Errorf(err.Error())
	}
	if allRepos[0].Name != "web" ||
		allRepos[0].Repository != "registry.digitalocean.com/<my-registry>/<my-image>" ||
		allRepos[0].Tag != "latest" {
		t.Errorf("error in unmarshalling input data")
	}

	checkForGitAndDockerHub(allRepos, &app)
	if app.Services[0].Name == "web" && app.Services[0].Git != nil {

		t.Errorf("error in checkForGitAndDockerHub")
	}

}
func TestFilterApps(t *testing.T) {
	test_input, err := ioutil.ReadFile("sample-golang.yaml")
	if err != nil {
		t.Errorf("error in reading test file")
	}
	var app godo.AppSpec
	err = yaml.Unmarshal(test_input, &app)
	if err != nil {
		t.Errorf("Error in unmarshalling test yaml")
	}
	if app.Services[0].Name == "web" && app.Services[0].Git.RepoCloneURL == "https://github.com/snormore/sample-golang.git" {
		t.Errorf("Error in parsing test data")
	}
	temp := `[ {
		"name": "web",
		"repository": "registry.digitalocean.com/<my-registry>/<my-image>",
		"tag": "latest"
	  }]`
	allRepos, err := getAllRepo(temp, "_")
	if err != nil {
		t.Errorf(err.Error())
	}
	if allRepos[0].Name != "web" ||
		allRepos[0].Repository != "registry.digitalocean.com/<my-registry>/<my-image>" ||
		allRepos[0].Tag != "latest" {
		t.Errorf("error in unmarshalling input data")
	}

	aErr := filterApps(allRepos, app)
	if aErr.name != "" {
		t.Errorf(aErr.name)
	}
	if app.Services[0].Image.RegistryType != "DOCR" ||
		app.Services[0].Image.Repository != "<my-image>" ||
		app.Services[0].Image.Tag != "latest" {
		t.Errorf("error in filterApps")
	}
}
func TestRetrieveAppId(t *testing.T) {
	appid, err := retrieveAppId("sample-golang")
	if appid == "" || err != nil {
		t.Errorf("Error in retrieving appid")
	}
	_, err = retrieveAppId("sadasfasfsa")
	if err == nil {
		t.Errorf("Not able to handle invalid name")
	}
}
