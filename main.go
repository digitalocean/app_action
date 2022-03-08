package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/digitalocean/app_action/internal/doctl"
	"github.com/digitalocean/app_action/internal/parser"
	parser_struct "github.com/digitalocean/app_action/internal/parser_struct"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

// AllError is used for handling errors
type AllError struct {
	name     string
	notFound []string
}

//go:generate mockgen -package main -source=main.go -self_package main -destination mock.go DoctlClient

//DoctlClient interface for doctl functions
type DoctlClient interface {
	ListDeployments(appID string) ([]godo.Deployment, error)
	RetrieveActiveDeploymentID(appID string) (string, error)
	RetrieveActiveDeployment(deploymentID string, appID string, input string) ([]parser_struct.UpdatedRepo, *godo.AppSpec, error)
	UpdateAppPlatformAppSpec(tmpfile, appID string) error
	CreateDeployments(appID string) error
	RetrieveFromDigitalocean() ([]godo.App, error)
	RetrieveAppID(appName string) (string, error)
	IsDeployed(appID string) error
	Deploy(input string, appName string) error
}

type action struct {
	appName   string
	images    string
	authToken string
	client    DoctlClient
}

func main() {
	//declaring variables for command line arguments input
	appName := os.Args[2]
	images := os.Args[1]
	authToken := os.Args[3]

	//check for authentication token
	if strings.TrimSpace(authToken) == "" {
		log.Fatal("No auth token provided")
	}

	//check for app name
	if strings.TrimSpace(appName) == "" {
		log.Fatal("No app name provided")
	}

	d, err := doctl.NewClient(authToken)
	if err != nil {
		log.Fatal(err)
	}

	a := &action{
		appName:   appName,
		images:    images,
		authToken: authToken,
		client:    &d,
	}

	err = a.run()
	if err != nil {
		log.Fatal(err)
	}
}

//run contains business logic of app_action
func (a *action) run() error {
	//redeploying app with the same app spec
	if strings.TrimSpace(a.images) == "" {
		err := a.client.Deploy(a.images, a.appName)
		if err != nil {
			return errors.Wrap(err, "triggering deploy")
		}
		return nil
	}

	//retrieve appID from users deployment
	appID, err := a.client.RetrieveAppID(a.appName)
	if err != nil {
		return errors.Wrap(err, "retrieving appID")
	}

	//retrieve deployment id of active deployment
	deploymentID, err := a.client.RetrieveActiveDeploymentID(appID)
	if err != nil {
		return errors.Wrap(err, "retrieving active deployment id")
	}

	//retrieve apps from deployment id
	input, apps, err := a.client.RetrieveActiveDeployment(deploymentID, appID, a.images)
	if err != nil {
		return errors.Wrap(err, "retrieving active deployment")
	}

	//updates local app spec based on user input
	tmpfile, err := a.updateLocalAppSpec(input, apps)
	if err != nil {
		return errors.Wrap(err, "updating local app spec")
	}

	// cleanup app spec file if exists after run
	defer func() {
		if _, err := os.Stat(tmpfile); err == nil {
			// deletes the local temp app spec file
			err = os.Remove(tmpfile)
			if err != nil {
				log.Fatalf("deleting local temp app spec file: %s", err)
			}
		}
	}()

	//updates app spec of the app using the local temp file and update
	err = a.client.UpdateAppPlatformAppSpec(tmpfile, appID)
	if err != nil {
		return errors.Wrap(err, "updating app spec")
	}

	//checks for deployment status
	err = a.client.IsDeployed(appID)
	if err != nil {
		return errors.Wrap(err, "checking deployment status")
	}

	return nil
}

//updateLocalAppSpec updates app spec based on users input and saves it in a local file called .do._app.yaml
func (a *action) updateLocalAppSpec(input []parser_struct.UpdatedRepo, appSpec *godo.AppSpec) (string, error) {
	//updates all the container images based on user input
	newErr := filterApps(input, *appSpec)
	if newErr.name != "" {
		log.Print(newErr.name)
		if len(newErr.notFound) != 0 {
			log.Fatalf("%v", newErr.notFound)
		}
		return "", errors.New(newErr.name)
	}

	//write to local temp file
	tmpfile, err := writeToTempFile(appSpec)
	if err != nil {
		return "", err
	}
	return tmpfile, nil
}

//writeToTempFile writes to a local temp file
func writeToTempFile(appSpec *godo.AppSpec) (string, error) {
	//parse App Spec to yaml
	newYaml, err := parser.ParseAppSpecToYaml(appSpec)
	if err != nil {
		return "", err
	}
	tmpfile, err := ioutil.TempFile("", "_do_app_*.yaml")
	if err != nil {
		return "", errors.Wrap(err, "Error in creating temp file")
	}
	if _, err := tmpfile.Write(newYaml); err != nil {
		tmpfile.Close()
		return "", errors.Wrap(err, "Error in writing to temp file")
	}
	if err := tmpfile.Close(); err != nil {
		return "", errors.Wrap(err, "Error in closing temp file")
	}
	return tmpfile.Name(), nil
}

// checkForGitAndDockerHub removes git, gitlab, github, DockerHub and DOCR images for the app name specified in the input json file
func checkForGitAndDockerHub(allFiles []parser_struct.UpdatedRepo, spec *godo.AppSpec) {
	//iterate through all the files of the input and save names in a map
	var nameMap = make(map[string]bool)
	for val := range allFiles {
		nameMap[allFiles[val].Name] = true
	}

	//remove git, gitlab, github and dockerhub spec of services with unique name declared in input
	for _, service := range spec.Services {
		if !nameMap[service.Name] {
			continue
		}
		service.Git = nil
		service.GitLab = nil
		service.GitHub = nil
		service.Image = nil
	}

	//remove git, gitlab, github and dockerhub spec of workers with unique name declared in input
	for _, worker := range spec.Workers {
		if !nameMap[worker.Name] {
			continue
		}
		worker.Git = nil
		worker.GitLab = nil
		worker.GitHub = nil
		worker.Image = nil
	}

	//remove git, gitlab, github and dockerhub spec of Jobs with unique name declared in input
	for _, job := range spec.Jobs {
		if !nameMap[job.Name] {
			continue
		}
		job.Git = nil
		job.GitLab = nil
		job.GitHub = nil
		job.Image = nil
	}
}

func makeImageSpec(updatedRepo parser_struct.UpdatedRepo) *godo.ImageSourceSpec {

	if updatedRepo.Image.RegistryType == "" {
		fmt.Println("::warning::Updating images without an ImageSourceSpec is deprecated. Please See: https://github.com/digitalocean/app_action/issues/10")
		repos := strings.Split(updatedRepo.Repository, `/`)
		repo := repos[len(repos)-1]
		return &godo.ImageSourceSpec{
			RegistryType: "DOCR",
			Repository:   repo,
			Tag:          updatedRepo.Tag,
		}
	}
	return &updatedRepo.Image
}

// filterApps filters git and DockerHub apps and then updates app spec with new ImageSourceSpec
func filterApps(allFiles []parser_struct.UpdatedRepo, appSpec godo.AppSpec) AllError {
	//remove all gitlab,github, git and dockerhub app info from appSpec for provided unique name component in input
	checkForGitAndDockerHub(allFiles, &appSpec)

	//iterate through all the files of the input and save names in a map
	var nameMap = make(map[string]bool)
	for val := range allFiles {
		nameMap[allFiles[val].Name] = true
	}

	//iterate through all services, worker and job to update AppSpec.ImageSourceSpec based on component name declared in input
	for key := range allFiles {
		for _, service := range appSpec.Services {
			if service.Name != allFiles[key].Name {
				continue
			}
			service.Image = makeImageSpec(allFiles[key])
			delete(nameMap, service.Name)
		}
		for _, worker := range appSpec.Workers {
			if worker.Name != allFiles[key].Name {
				continue
			}

			worker.Image = makeImageSpec(allFiles[key])
			delete(nameMap, worker.Name)
		}
		for _, job := range appSpec.Jobs {
			if job.Name != allFiles[key].Name {
				continue
			}

			job.Image = makeImageSpec(allFiles[key])
			delete(nameMap, job.Name)
		}

		//if functions component unique name is mentioned in the user input throw error as functions components do not support containers
		for _, functions := range appSpec.Functions {
			if functions.Name != allFiles[key].Name {
				continue
			}

			return AllError{
				name: fmt.Sprintf("Functions components in App Platform do not support containers: %s", functions.Name),
			}
		}
		//if static sites unique name is mentioned in the user input throw error as static sites do not support containers
		for _, static := range appSpec.StaticSites {
			if static.Name != allFiles[key].Name {
				continue
			}

			return AllError{
				name: fmt.Sprintf("Static sites in App Platform do not support containers: %s", static.Name),
			}
		}
	}

	if len(nameMap) == 0 {
		return AllError{}
	}

	keys := make([]string, 0, len(nameMap))

	for k := range nameMap {
		keys = append(keys, k)
	}

	return AllError{
		name:     "all components with following names were not found in your deployed app spec",
		notFound: keys,
	}
}
