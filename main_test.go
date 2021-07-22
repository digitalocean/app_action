package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/digitalocean/godo"
	"gopkg.in/yaml.v2"
)

type MockDoctlDependencies struct {
	dep doctlDependencies
}

func (m *MockDoctlDependencies) isAuthenticated(name string, token string) error {
	return nil
}
func (m *MockDoctlDependencies) isDeployed(appID string) error {
	fmt.Println("Build successful")
	return nil
}
func (m *MockDoctlDependencies) getAllRepo(input string, appName string) ([]UpdatedRepo, error) {
	if strings.TrimSpace(input) == "" {
		appID, err := m.retrieveAppID(appName)
		if err != nil {
			return nil, err
		}
		err = m.isDeployed(appID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	var allRepos []UpdatedRepo
	err := json.Unmarshal([]byte(input), &allRepos)
	if err != nil {
		return nil, errors.New("error in parsing json data from file")
	}
	return allRepos, nil
}
func (m *MockDoctlDependencies) retrieveActiveDeployment(appID string) ([]byte, error) {

	output, err := ioutil.ReadFile("/testdata/temp")
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (m *MockDoctlDependencies) updateAppPlatformAppSpec(appID string, appSpec string) error {
	return nil
}
func (m *MockDoctlDependencies) retrieveAppID(name string) (string, error) {
	return "5e6b7bd1-d04e-4694-8679-bf8651f72663", nil
}
func TestGetAllRepo(t *testing.T) {
	temp := `[ {
		"name": "frontend",
		"repository": "registry.digitalocean.com/<my-registry>/<my-image>",
		"tag": "latest"
	  }]`

	var test1Interface doctlDependencies
	t1 := MockDoctlDependencies{test1Interface}
	allRepos, err := t1.getAllRepo(temp, "_")
	if err != nil {
		t.Errorf("Error in parsing input json data")
	}
	if allRepos[0].Name != "frontend" ||
		allRepos[0].Repository != "registry.digitalocean.com/<my-registry>/<my-image>" ||
		allRepos[0].Tag != "latest" {
		t.Errorf("Error in unmarshal")
	}
	//testing individual deployment for get all repo
	_, err = t1.getAllRepo("", "TEST_APP_NAME")
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestCheckForGitAndDockerHub(t *testing.T) {
	testInput, err := ioutil.ReadFile("testdata/sample-golang.yaml")
	if err != nil {
		t.Errorf("error in reading test file")
	}
	var app godo.AppSpec
	err = yaml.Unmarshal(testInput, &app)
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
	var test2Interface doctlDependencies
	t2 := MockDoctlDependencies{test2Interface}
	allRepos, err := t2.getAllRepo(temp, "_")
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
	testInput, err := ioutil.ReadFile("testdata/sample-golang.yaml")
	if err != nil {
		t.Errorf("error in reading test file")
	}
	var app godo.AppSpec
	err = yaml.Unmarshal(testInput, &app)
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
	var test2Interface doctlDependencies
	t3 := MockDoctlDependencies{test2Interface}
	allRepos, err := t3.getAllRepo(temp, "_")
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

// func TestRetrieveAppID(t *testing.T) {
// 	var test2Interface doctlDependencies
// 	t4 := DoctlServices{test2Interface}
// 	appID, err := t4.retrieveAppID("sample-golang")
// 	if appID == "" || err != nil {
// 		t.Errorf("Error in retrieving appid")
// 	}
// 	_, err = t4.retrieveAppID("sadasfasfsa")
// 	if err == nil {
// 		t.Errorf("Not able to handle invalid name")
// 	}
// }

// func TestIsAuthenticated(t *testing.T) {
// 	var test2Interface doctlDependencies
// 	t2 := DoctlServices{test2Interface}
// 	err := t2.isAuthenticated(os.Getenv("TOKEN"))
// 	if err != nil {
// 		t.Errorf("Error in isAuthenticated %s", err)
// 	}
// }
