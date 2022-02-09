package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/digitalocean/app_action/internal/parser"
	"github.com/digitalocean/godo"
	gomock "github.com/golang/mock/gomock"
	"gopkg.in/yaml.v2"
)

//TestParseJsonInput uses custom input to check if the parseJsonInput function is working properly
func TestParseJsonInput(t *testing.T) {
	temp := `[ {
		"name": "frontend",
		"repository": "registry.digitalocean.com/<my-registry>/<my-image>",
		"tag": "latest"
	  }]`
	allRepos, err := parser.ParseJsonInput(temp)
	if err != nil {
		t.Errorf("Error in parsing input json data")
	}
	if allRepos[0].Name != "frontend" ||
		allRepos[0].Repository != "registry.digitalocean.com/<my-registry>/<my-image>" ||
		allRepos[0].Tag != "latest" {
		t.Errorf("Error in unmarshal")
	}
}

//TestCheckForGitAndDockerHub uses custom input to check if the checkForGitAndDockerHub is working
func TestCheckForGitAndDockerHub(t *testing.T) {
	//sample-golang is app spec used for testing purposes
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
	allRepos, err := parser.ParseJsonInput(temp)
	if err != nil {
		t.Errorf(err.Error())
	}
	if allRepos[0].Name != "web" ||
		allRepos[0].Repository != "registry.digitalocean.com/<my-registry>/<my-image>" ||
		allRepos[0].Tag != "latest" {
		t.Errorf("error in unmarshalling input data")
	}

	//check for git,github,gitlab,DOCR and dockerhub removal for app name provided in user input
	checkForGitAndDockerHub(allRepos, &app)
	if app.Services[0].Name == "web" && app.Services[0].Git != nil {

		t.Errorf("error in checkForGitAndDockerHub")
	}

}

//TestFilterApps tests filterApps function using testdata/sample-golang.yaml as input
func TestFilterApps(t *testing.T) {
	//sample-golang is app spec used for testing purposes
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

	//paseJsonInput function is used to parse the input json data
	allRepos, err := parser.ParseJsonInput(temp)
	if err != nil {
		t.Errorf(err.Error())
	}
	if allRepos[0].Name != "web" ||
		allRepos[0].Repository != "registry.digitalocean.com/<my-registry>/<my-image>" ||
		allRepos[0].Tag != "latest" {
		t.Errorf("error in unmarshalling input data")
	}

	//filterApps function is used to filter the app spec based on the app name provided in user input
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

//TestUpdateLocalAppSpec tests all the non doctl dependent functions
func TestUpdateLocalAppSpec(t *testing.T) {
	t1Input := `[{
		  "name": "web",
		  "repository": "registry.digitalocean.com/sample-go/add_sample",
		  "tag": "latest"
		}
	  ]`

	//temp is the deployment spec scraped from actual deployment used for testing purposes
	testInput, err := ioutil.ReadFile("testdata/temp")
	if err != nil {
		t.Errorf("error in reading test file")
	}

	a := &action{
		appName: "sample-golang",
		images:  t1Input,
	}

	allRepos, err := parser.ParseJsonInput(t1Input)
	if err != nil {
		t.Errorf(err.Error())
	}
	//parseDeploymentSpec
	appSpec, err := parser.ParseDeploymentSpec(testInput)
	if err != nil {
		t.Errorf(err.Error())
	}

	//test for all functions which are independent of doctl
	file, err := a.updateLocalAppSpec(allRepos, appSpec[0].Spec)
	if err != nil {
		t.Errorf(err.Error())
	}
	f1, err1 := ioutil.ReadFile(file)
	if err1 != nil {
		log.Fatal(err1)
	}

	//read updatedAppSpec.yaml to compare the final output with expected output
	f2, err2 := ioutil.ReadFile("testdata/updatedAppSpec.yaml")
	if err2 != nil {
		log.Fatal(err2)
	}
	if bytes.Equal(f1, f2) == false {
		t.Errorf("error in parsing app spec yaml file")
	}
	os.Remove(file)
}

func Test_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	appID := "2a91c9e3-253f-4c75-99e5-b81b9c3f744f"
	activeDeploymentID := "fac38395-30f3-4c59-9e6c-3a67523f51de"
	sampleImages := `[{
		"name": "web",
		"repository": "registry.digitalocean.com/sample-go/add_sample",
		"tag": "latest"
	  }
	]`

	//parse input data
	sampleImagesRepo, err := parser.ParseJsonInput(sampleImages)
	if err != nil {
		t.Errorf(err.Error())
	}

	do := NewMockDoctlClient(ctrl)
	do.EXPECT().RetrieveAppID(gomock.Eq("sample-golang")).Return(appID, nil)
	do.EXPECT().RetrieveActiveDeploymentID(gomock.Eq(appID)).Return(activeDeploymentID, nil)
	//temp is the deployment spec scraped from actual deployment used for testing purposes
	testInput, err := ioutil.ReadFile("testdata/temp")
	if err != nil {
		t.Errorf("error in reading test file")
	}
	//parse testInput data
	deployments, err := parser.ParseDeploymentSpec(testInput)
	if err != nil {
		t.Errorf(err.Error())
	}
	do.EXPECT().RetrieveActiveDeployment(gomock.Eq(activeDeploymentID), gomock.Eq(appID), gomock.Eq(sampleImages)).Return(sampleImagesRepo, deployments[0].Spec, nil)
	do.EXPECT().UpdateAppPlatformAppSpec(gomock.Any(), appID).Return(nil)

	do.EXPECT().IsDeployed(appID).Return(nil)

	a := &action{
		appName: "sample-golang",
		images:  sampleImages,
		client:  do,
	}

	err = a.run()
	if err != nil {
		t.Fail()
	}
}
