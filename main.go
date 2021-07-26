package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
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

type doctlDependencies interface {
	retrieveActiveDeployment(appID string) ([]byte, error)
	retrieveAppID(appName string) (string, error)
	isDeployed(appID string) error
	updateAppPlatformApp(appID string) error
	isAuthenticated(token string) error
	createDeployments(appID string) error
}
type DoctlServices struct {
	dep doctlDependencies
}

func main() {
	var dependent doctlDependencies
	d := DoctlServices{dep: dependent}
	//user authentication
	err := d.isAuthenticated(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	input, err := d.getAllRepo(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	//retrieve AppID from users deployment
	appID, err := d.retrieveAppID(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	//retrieve id of active deployment
	apps, err := d.retrieveActiveDeployment(appID)
	if err != nil {
		log.Fatal(err)
	}
	var app []godo.App
	err = json.Unmarshal(apps, &app)
	if err != nil {
		log.Fatal("Error in retrieving app spec: ", err)
	}
	appSpec := *app[0].Spec

	//updates all the docr images based on user input
	newErr := filterApps(input, appSpec)
	if newErr.name != "" {
		log.Print(newErr.name)
		if len(newErr.notFound) != 0 {
			log.Fatalf("%v", newErr.notFound)
		}
		os.Exit(1)
	}
	newYaml, err := yaml.Marshal(appSpec)
	if err != nil {
		log.Fatal("Error in building spec from json data")
	}
	//write to local file
	err = ioutil.WriteFile(".do._app.yaml", newYaml, 0644)
	if err != nil {
		log.Fatal("Error in writing to yaml")
	}

	//updates app spec of the app and deploys it
	err = d.updateAppPlatformAppSpec(appID)
	if err != nil {
		log.Fatal(err)
	}
	//checks for deployment status
	err = d.isDeployed(appID)
	if err != nil {
		log.Fatal(err)
	}
}

//isAuthenticated checks for authentication
func (d *DoctlServices) isAuthenticated(token string) error {
	val, err := exec.Command("sh", "-c", fmt.Sprintf("doctl auth init --access-token %s", token)).Output()
	if err != nil {
		return fmt.Errorf("unable to authenticate user: %s", val)
	}
	return nil
}

//getCurrentDeployment returns the current deployment
func (d *DoctlServices) getCurrentDeployment(appID string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app list-deployments %s -ojson", appID))
	spec, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error in retrieving list of deployments")
	}
	return spec, nil

}

//isDeployed checks for the status of deployment
func (d *DoctlServices) isDeployed(appID string) error {
	done := false
	for !done {
		fmt.Println("App Platform is Building ....")
		spec, err := d.getCurrentDeployment(appID)
		if err != nil {
			return errors.Wrap(err, "error in retrieving list of deployments")
		}
		var app []godo.Deployment
		err = json.Unmarshal(spec, &app)
		if err != nil {
			return errors.Wrap(err, "error in parsing deployment")
		}
		if app[0].Phase == "ACTIVE" {
			fmt.Println("Build successful")
			return nil
		}
		if app[0].Phase == "Failed" {
			fmt.Println("Build unsuccessful")
			return errors.Wrap(err, "build unsuccessful")
		}
	}
	return nil
}

//retrieveActiveDeployment retrieves currently deployed app spec of on App Platform app
func (d *DoctlServices) retrieveActiveDeployment(appID string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl apps get --format ActiveDeployment.ID --no-header %s", appID))
	deployID, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve active deployment")
	}
	deploymentID := strings.TrimSpace(string(deployID))

	//get app based on appID
	cmd = exec.Command("sh", "-c", fmt.Sprintf("doctl app get-deployment %s %s -ojson", appID, string(deploymentID)))
	apps, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error in retrieving currently deployed app id")
	}
	return apps, nil
}

//updateAppPlatformAppSpec updates app spec and creates deployment for the app
func (d *DoctlServices) updateAppPlatformAppSpec(appID string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app update %s --spec .do._app.yaml", appID))
	_, err := cmd.Output()
	if err != nil {
		fmt.Print(err)
		return errors.Wrap(err, "unable to update app")
	}
	err = d.createDeployments(appID)
	if err != nil {
		return err
	}
	return nil
}

//createDeployments creates deployment for the app
func (d *DoctlServices) createDeployments(appID string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app create-deployment %s", appID))
	_, err := cmd.Output()
	if err != nil {
		return errors.Wrap(err, "unable to create-deployment for app")
	}
	return nil
}

// getAllRepo reads the file and return json object of type UpdatedRepo
func (d *DoctlServices) getAllRepo(input string, appName string) ([]UpdatedRepo, error) {
	//takes care of empty input for Deployment
	if strings.TrimSpace(string(input)) == "" {
		appID, err := d.retrieveAppID(appName)
		if err != nil {
			return nil, err
		}
		err = d.createDeployments(appID)
		if err != nil {
			return nil, err
		}
		error := d.isDeployed(appID)
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
		return nil, errors.Wrap(err, "error in parsing json data from file")
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

// retrieveAppID retrieves app id from app platform
func (d *DoctlServices) retrieveAppID(appName string) (string, error) {
	cmd := exec.Command("sh", "-c", "doctl app list -ojson")
	apps, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "unable to get user app data from digitalocean")
	}

	//parsing incoming data for AppId
	var arr []godo.App
	err = json.Unmarshal(apps, &arr)
	if err != nil {
		return "", errors.Wrap(err, "error in parsing data for AppId")
	}
	var appID string

	for k := range arr {
		if arr[k].Spec.Name == appName {
			appID = arr[k].ID
			break
		}
	}
	if appID == "" {
		return "", errors.Wrap(err, "app not found")
	}

	return appID, nil
}
