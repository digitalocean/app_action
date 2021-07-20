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
	"sigs.k8s.io/yaml"
)

//used for parsing json object of changed repo
type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}

type AllError struct {
	name     string
	notFound []string
}

func isDeployed(appId string) error {
	done := false
	for !done {
		fmt.Println("App Platform is Building ....")
		cmd := exec.Command("sh", "-c", `doctl app list-deployments `+appId+` -ojson `)
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
			return error.New("Build unsuccessful")
		}
	}
	return nil
}

//reads the file and return json object of type UpdatedRepo
func getAllRepo(input string, appName string) ([]UpdatedRepo, error) {

	//takes care of empty input for  Deployment
	if strings.TrimSpace(string(input)) == "" {
		appId, err := retrieveAppId(appName)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command("sh", "-c", `doctl app create-deployment `+appId)
		_, err = cmd.Output()
		if err != nil {
			return nil, errors.New("unable to create-deployment for app")
		}
		error := isDeployed(appId)
		if error != nil {
			return nil, error
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

//Remove git and DockerHub
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

//filters git and DockerHub apps and then updates app spec with DOCR
func filterApps(allFiles []UpdatedRepo, appSpec godo.AppSpec) AllError {
	checkForGitAndDockerHub(allFiles, &appSpec)
	var nameMap = make(map[string]bool)
	for val := range allFiles {
		nameMap[allFiles[val].Name] = true
	}

	for key, _ := range allFiles {
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
					name: `Static sites in App Platform do not support DOCR: ` + static.Name,
				}
			}
		}

	}
	if len(nameMap) == 0 {
		return AllError{}
	} else {
		keys := make([]string, 0, len(nameMap))
		for k := range nameMap {
			keys = append(keys, k)
		}
		error_new := AllError{
			name:     "all files not found",
			notFound: keys,
		}
		return error_new
	}

}

// func uploadToDOCR(data []UpdatedRepo) error {

// 	for k, _ := range data {
// 		cmd := exec.Command("sh", "-c", `doctl registry login`)
// 		_, err := cmd.Output()
// 		if err != nil {
// 			log.Fatal("Unable to login to registry:", data[k].Name)
// 			return err
// 		}
// 		if data[k].Tag != "" && data[k].Name != "" && data[k].Repository != "" {
// 			cmd := exec.Command("sh", "-c", `docker push `+data[k].Repository+`:`+data[k].Tag)
// 			_, err := cmd.Output()
// 			if err != nil {
// 				log.Fatal("Unable to upload image to docr app:", data[k].Name)
// 				return err
// 			}
// 		} else if data[k].Name != "" && data[k].Repository != "" && data[k].Tag == "" {
// 			cmd := exec.Command("sh", "-c", `docker push `+data[k].Repository+`:latest`)
// 			_, err := cmd.Output()
// 			if err != nil {
// 				log.Fatal("Unable to upload image to docr app:", data[k].Name)
// 				return err
// 			}
// 		}

// 	}
// 	return nil

// }
func retrieveAppId(appName string) (string, error) {
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
	var appId string

	for k := range arr {
		if arr[k].Spec.Name == appName {
			appId = arr[k].ID
			break
		}
	}
	if appId == "" {
		return "", errors.New("app not found")
	}
	return appId, nil
}
func main() {
	//retrieve input
	fmt.Println(os.Args[1])
	name := os.Args[2]
	fmt.Println("this is name", name)
	//doctl
	_, err := exec.Command("sh", "-c", `doctl auth init --access-token `+os.Args[3]).Output()
	if err != nil {
		log.Fatal("Unable to authenticate ", err.Error())
		os.Exit(1)
	}
	//read json file from input
	input, err := getAllRepo(os.Args[1], name)
	if err != nil {
		log.Fatal("Error in Retrieving json data: ", err)
		os.Exit(1)
	}

	//retrieve AppId from users deployment
	appId, err := retrieveAppId(name)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	//retrieve deployment id
	cmd := exec.Command("sh", "-c", "doctl apps get --format ActiveDeployment.ID --no-header "+appId)
	deployId, err := cmd.Output()
	if err != nil {
		log.Fatal("Unable to retrieve active deployment:", err)
		os.Exit(1)
	}
	deploymentId := strings.TrimSpace(string(deployId))
	//get app based on appID
	cmd = exec.Command("sh", "-c", `doctl app get-deployment `+appId+` `+string(deploymentId)+` -ojson`)
	apps, err := cmd.Output()
	if err != nil {
		log.Fatal("Unable to retrieve currently deployed app id:", err)
		os.Exit(1)
	}
	var app []godo.App
	err = json.Unmarshal(apps, &app)
	if err != nil {
		log.Fatal("Error in retrieving app spec: ", err)
		os.Exit(1)
	}
	appSpec := *app[0].Spec
	//docr registry login
	cmd = exec.Command("sh", "-c", "doctl registry login")
	_, err = cmd.Output()
	if err != nil {
		log.Fatal("Unable to login to digitalocean registry:", err)
		os.Exit(1)
	}
	//docr registry upload
	// err = uploadToDOCR(input)
	// if err != nil {
	// 	log.Fatal("DOCR update error occurred")
	// 	os.Exit(1)
	// }
	//updates all the docr images based on users input
	new_err := filterApps(input, appSpec)
	if new_err.name != "" {
		log.Fatal(new_err.name)
		if len(new_err.notFound) != 0 {
			log.Fatalf("%v", new_err.notFound)
		}
		os.Exit(1)
	}
	newYaml, err := yaml.Marshal(appSpec)
	if err != nil {
		log.Fatal("Error in building spec from json data")
		os.Exit(1)
	}
	err = ioutil.WriteFile(".do._app.yaml", newYaml, 0644)
	if err != nil {
		log.Fatal("Error in writing to yaml")
		os.Exit(1)
	}
	cmd = exec.Command("sh", "-c", `doctl app update `+appId+` --spec .do._app.yaml`)
	_, err = cmd.Output()
	if err != nil {
		log.Fatal("Unable to update app:", err)
		os.Exit(1)
	}

	cmd = exec.Command("sh", "-c", `doctl app create-deployment `+appId)
	_, err = cmd.Output()
	if err != nil {
		log.Fatal("Unable to create-deployment for app:", err)
		os.Exit(1)
	}
	isDeployed(appId)

}
