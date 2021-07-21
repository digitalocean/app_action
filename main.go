package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/digitalocean/godo"
	"gopkg.in/yaml.v2"
)

// UpdatedRepo used for parsing json object of changed repo
type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}

// AllError is used for handling errors
type AllError struct {
	name     string
	notFound []string
}

func main() {
	//retrieve input
	name := os.Args[2]
	//doctl

	_, err := exec.Command("sh", "-c", fmt.Sprintf("doctl auth init --access-token %s", os.Args[3])).Output()
	if err != nil {
		log.Fatal("Unable to authenticate ", err.Error())
	}
	//read json file from input
	input, err := getAllRepo(os.Args[1], name)
	if err != nil {
		log.Fatal("Error in Retrieving json data: ", err)
	}

	//retrieve AppID from users deployment
	appID, err := retrieveAppID(name)
	if err != nil {
		log.Fatal(err)
	}

	//retrieve deployment id
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl apps get --format ActiveDeployment.ID --no-header %s", appID))
	deployID, err := cmd.Output()
	if err != nil {
		log.Fatal("Unable to retrieve active deployment:", err)
	}
	deploymentID := strings.TrimSpace(string(deployID))

	//get app based on appID
	cmd = exec.Command("sh", "-c", fmt.Sprintf("doctl app get-deployment %s %s -ojson", appID, string(deploymentID)))
	apps, err := cmd.Output()
	if err != nil {
		log.Fatal("Unable to retrieve currently deployed app id:", err)
	}

	var app []godo.App
	err = json.Unmarshal(apps, &app)
	if err != nil {
		log.Fatal("Error in retrieving app spec: ", err)
	}
	appSpec := *app[0].Spec

	//docr registry login
	cmd = exec.Command("sh", "-c", "doctl registry login")
	_, err = cmd.Output()
	if err != nil {
		log.Fatal("Unable to login to digitalocean registry:", err)
	}

	//updates all the docr images based on users input
	newErr := filterApps(input, appSpec)
	if newErr.name != "" {
		log.Fatal(newErr.name)
		if len(newErr.notFound) != 0 {
			log.Fatalf("%v", newErr.notFound)
		}
		os.Exit(1)
	}

	newYaml, err := yaml.Marshal(appSpec)
	if err != nil {
		log.Fatal("Error in building spec from json data")
	}

	err = ioutil.WriteFile(".do._app.yaml", newYaml, 0644)
	if err != nil {
		log.Fatal("Error in writing to yaml")
	}

	cmd = exec.Command("sh", "-c", fmt.Sprintf("doctl app update %s --spec .do._app.yaml", appID))
	_, err = cmd.Output()
	if err != nil {
		log.Fatal("Unable to update app:", err)
	}

	cmd = exec.Command("sh", "-c", fmt.Sprintf("doctl app create-deployment %s", appID))
	_, err = cmd.Output()
	if err != nil {
		log.Fatal("Unable to create-deployment for app:", err)
	}
	isDeployed(appID)
}
func isDeployed(appID string) error {
	done := false
	for !done {
		fmt.Println("App Platform is Building ....")
		cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app get-deployment %s -ojson", appID))
		spec, err := cmd.Output()
		if err != nil {
			return errors.New("error in retrieving list of deployments")
		}
		var app []godo.Deployment
		err = json.Unmarshal(spec, &app)
		if err != nil {
			return errors.New("error in parsing deployment")
		}
		if app[0].Phase == "ACTIVE" {
			fmt.Println("Build successful")
			return nil
		}
		if app[0].Phase == "Failed" {
			fmt.Println("Build unsuccessful")
			return errors.New("build unsuccessful")
		}
	}
	return nil
}

// getAllRepo reads the file and return json object of type UpdatedRepo
func getAllRepo(input string, appName string) ([]UpdatedRepo, error) {

	//takes care of empty input for  Deployment
	if strings.TrimSpace(string(input)) == "" {
		appID, err := retrieveAppID(appName)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command("sh", "-c", "doctl app create-deployment %s", appID)
		_, err = cmd.Output()
		if err != nil {
			return nil, errors.New("unable to create-deployment for app")
		}
		error := isDeployed(appID)
		if error != nil {
			return nil, error
		}
		if error == nil {
			os.Exit(0)
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

// checkForGitAndDockerHub Remove git and DockerHub
func checkForGitAndDockerHub(allFiles []UpdatedRepo, spec *godo.AppSpec) {
	var nameMap = make(map[string]bool)
	for val := range allFiles {
		nameMap[allFiles[val].Name] = true
	}
	for _, service := range spec.Services {
		if !nameMap[service.Name] {
			continue
		}
		service.Git = nil
		service.GitLab = nil
		service.GitHub = nil
		service.Image = nil
	}
	for _, worker := range spec.Workers {
		if !nameMap[worker.Name] {
			continue
		}
		worker.Git = nil
		worker.GitLab = nil
		worker.GitHub = nil
		worker.Image = nil
	}
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

// filterApps filters git and DockerHub apps and then updates app spec with DOCR
func filterApps(allFiles []UpdatedRepo, appSpec godo.AppSpec) AllError {
	checkForGitAndDockerHub(allFiles, &appSpec)
	var nameMap = make(map[string]bool)
	for val := range allFiles {
		nameMap[allFiles[val].Name] = true
	}

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

// retrieveAppID ...
func retrieveAppID(appName string) (string, error) {
	cmd := exec.Command("sh", "-c", "doctl app list -ojson")
	apps, err := cmd.Output()
	if err != nil {
		return "", errors.New("unable to get user app data from digitalocean")
	}

	//parsing incoming data for AppId
	var arr []godo.App
	err = json.Unmarshal(apps, &arr)
	if err != nil {
		return "", errors.New("error in parsing data for AppId")
	}
	var appID string

	for k := range arr {
		if arr[k].Spec.Name == appName {
			appID = arr[k].ID
			break
		}
	}
	if appID == "" {
		return "", errors.New("app not found")
	}

	return appID, nil
}
