package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	mylib "github.com/ParamPatel207/app_action/internal/doctl"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// AllError is used for handling errors
type AllError struct {
	name     string
	notFound []string
}

func main() {
	//declaring variables for command line arguments input
	appName := os.Args[2]
	listOfImage := os.Args[1]
	authToken := os.Args[3]

	//declaring interface for doctl functions
	var dependent mylib.DoctlDependencies
	d := mylib.DoctlServices{Dep: dependent}

	//check for auth token
	if strings.TrimSpace(authToken) == "" {
		log.Fatal("No auth token provided")
	}

	//check for app name
	if strings.TrimSpace(appName) == "" {
		log.Fatal("No app name provided")
	}

	//redeploying app with same app spec
	if strings.TrimSpace(listOfImage) == "" {
		err := d.ReDeploy(listOfImage, appName)
		if err != nil {
			log.Fatal(err)
		}
	}
	//run functional logic of the code
	run(appName, listOfImage, authToken, &d)

}

func run(appName, listOfImage, authToken string, d *mylib.DoctlServices) {
	//user authentication
	err := d.IsAuthenticated(authToken)
	if err != nil {
		log.Fatal(err)
	}

	//retrieve AppID from users deployment
	appID, err := d.RetrieveAppID(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	//retrieve id of active deployment
	deploymentID, err := d.RetrieveActiveDeploymentID(appID)
	if err != nil {
		log.Fatal(err)
	}

	//retrieve apps from deployment id
	apps, err := d.RetrieveActiveDeployment(deploymentID, appID)
	if err != nil {
		log.Fatal(err)
	}

	//update local app spec
	err = updateLocalAppSpec(listOfImage, appName, apps)
	if err != nil {
		log.Fatal(err)
	}

	//updates app spec of the app using the local temp file and update
	err = d.UpdateAppPlatformAppSpec(appID)
	if err != nil {
		log.Fatal(err)
	}

	//Create a new deployment from the updated app spec
	err = d.CreateDeployments(appID)
	if err != nil {
		log.Fatal(err)
	}

	//checks for deployment status
	err = d.IsDeployed(appID)
	if err != nil {
		log.Fatal(err)
	}

	//deletes the local temp app spec file
	err = os.Remove(".do._app.yaml")
	if err != nil {
		log.Fatal(err, "Error in removing local file")
	}
}

func updateLocalAppSpec(listOfImage string, appName string, apps []byte) error {
	//parse array of input objects
	input, err := parseJsonInput(listOfImage, appName)
	if err != nil {
		return err
	}

	//parse array of Deployment objects
	appSpec, err := parseDeploymentSpec(apps)
	if err != nil {
		return err
	}

	//updates all the docr images based on user input
	newErr := filterApps(input, *appSpec)
	if newErr.name != "" {
		log.Print(newErr.name)
		if len(newErr.notFound) != 0 {
			log.Fatalf("%v", newErr.notFound)
		}
		return errors.New(newErr.name)
	}

	//build yaml from the input json data
	newYaml, err := yaml.Marshal(appSpec)
	if err != nil {
		return errors.Wrap(err, "Error in building yaml")
	}

	//write to local temp file
	err = ioutil.WriteFile(".do._app.yaml", newYaml, 0644)
	if err != nil {
		return errors.Wrap(err, "Error in writing local yaml file")
	}
	return nil

}

// parseDeploymentSpec parses deployment array and retrieves appSpec of recent deployment
func parseDeploymentSpec(apps []byte) (*godo.AppSpec, error) {
	var app []godo.App
	err := json.Unmarshal(apps, &app)
	if err != nil {
		log.Fatal("Error in retrieving app spec: ", err)
	}
	appSpec := *app[0].Spec
	return &appSpec, nil
}

// checkForGitAndDockerHub Remove git and DockerHub
func checkForGitAndDockerHub(allFiles []mylib.UpdatedRepo, spec *godo.AppSpec) {
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

// parseJsonInput takes the array of json object as input and unique name of users app as appName
//it parses the input and returns UpdatedRepo of the input
func parseJsonInput(input string, appName string) ([]mylib.UpdatedRepo, error) {
	//takes care of empty json Deployment (use case where we redeploy using same app spec)
	var allRepos []mylib.UpdatedRepo
	err := json.Unmarshal([]byte(input), &allRepos)
	if err != nil {
		return nil, errors.Wrap(err, "error in parsing json data from file")
	}
	return allRepos, nil
}

// filterApps filters git and DockerHub apps and then updates app spec with DOCR
func filterApps(allFiles []mylib.UpdatedRepo, appSpec godo.AppSpec) AllError {
	//remove all gitlab,github, git and dockerhub app info from appSpec for provided unique name component
	//in input
	checkForGitAndDockerHub(allFiles, &appSpec)
	//iterate through all the files of the input and save names in a map
	var nameMap = make(map[string]bool)
	for val := range allFiles {
		nameMap[allFiles[val].Name] = true
	}
	//iterate through all services, worker and job to update DOCR image in AppSpec based on unique name declared in input
	for key := range allFiles {
		for _, service := range appSpec.Services {
			if service.Name != allFiles[key].Name {
				continue
			} else {
				repos := strings.Split(allFiles[key].Repository, `/`)
				repo := repos[len(repos)-1]
				service.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: repo, Tag: allFiles[key].Tag}
				delete(nameMap, service.Name)
			}
		}
		for _, worker := range appSpec.Workers {
			if worker.Name != allFiles[key].Name {
				continue
			}
			repos := strings.Split(allFiles[key].Repository, `/`)
			repo := repos[len(repos)-1]
			worker.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: repo, Tag: allFiles[key].Tag}
			delete(nameMap, worker.Name)
		}
		for _, job := range appSpec.Jobs {
			if job.Name != allFiles[key].Name {
				continue
			}
			repos := strings.Split(allFiles[key].Repository, `/`)
			repo := repos[len(repos)-1]
			job.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: repo, Tag: allFiles[key].Tag}
			delete(nameMap, job.Name)
		}
		//if static site name is mentioned throw error as static sites do not support DOCR
		for _, static := range appSpec.StaticSites {
			if static.Name != allFiles[key].Name {
				continue
			} else {
				return AllError{
					name: fmt.Sprintf("Static sites in App Platform do not support DOCR: %s", static.Name),
				}
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
		name:     "all files not found",
		notFound: keys,
	}
}
