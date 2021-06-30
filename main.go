package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/digitalocean/godo"
	"gopkg.in/yaml.v2"
)

//used for parsing json object of changed repo
type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}

type AppSpec struct {
	// The name of the app. Must be unique across all apps in the same account.
	Name string `json:"name"`
	// Workloads which expose publicy-accessible HTTP services.
	Services []*godo.AppServiceSpec `json:"services,omitempty"`
	// Content which can be rendered to static web assets.
	Static_sites []*godo.AppStaticSiteSpec `json:"static_sites,omitempty"`
	// Workloads which do not expose publicly-accessible HTTP services.
	Workers []*godo.AppWorkerSpec `json:"workers,omitempty"`
	// Pre and post deployment workloads which do not expose publicly-accessible HTTP routes.
	Jobs []*godo.AppJobSpec `json:"jobs,omitempty"`
	// Database instances which can provide persistence to workloads within the application.
	Databases []*godo.AppDatabaseSpec `json:"databases,omitempty"`
	// A set of hostnames where the application will be available.
	Domains []*godo.AppDomainSpec `json:"domains,omitempty"`
	Region  string                `json:"region,omitempty"`
	// A list of environment variables made available to all components in the app.
	Envs []*godo.AppVariableDefinition `json:"envs,omitempty"`
}

//reads the file from fileLocation
func readFileFrom(fileLocation string) ([]byte, error) {
	byteValue, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		log.Fatal("Error in reading from file: ", err)
		return []byte{}, err
	}
	return byteValue, err
}

//reads the file and return json object of type UpdatedRepo
func getAllRepo(input string, appSpec string) ([]UpdatedRepo, AppSpec, error) {
	//parsing input
	jsonByteValue, err := readFileFrom(input)
	if err != nil {
		log.Fatal("Error in reading from file: ", err)
		return nil, AppSpec{}, err
	}
	var allRepos []UpdatedRepo
	err = json.Unmarshal(jsonByteValue, &allRepos)
	if err != nil {
		log.Fatal("Error in parsing json data from file: ", err)
		return nil, AppSpec{}, err
	}
	//parsing yml
	yamlByteValue, err := readFileFrom(appSpec)
	if err != nil {
		log.Fatal("Error in reading from file: ", err)
		return nil, AppSpec{}, err
	}
	var spec AppSpec
	err = yaml.Unmarshal(yamlByteValue, &spec)
	if err != nil {
		return nil, AppSpec{}, err
	}
	fmt.Printf("%+v\n", spec)
	return allRepos, spec, nil

}
func checkForGitAndDockerHub(allFiles []UpdatedRepo, spec *AppSpec) {
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

func execCommand(allFiles []UpdatedRepo, appSpec AppSpec) {
	checkForGitAndDockerHub(allFiles, &appSpec)
	var nameMap = make(map[string]bool)
	for val := range allFiles {
		nameMap[allFiles[val].Name] = true
	}

	for key, _ := range allFiles {
		for _, service := range appSpec.Services {
			if service.Name != allFiles[key].Name {
				continue
			}
			service.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: allFiles[key].Repository, Tag: allFiles[key].Tag}
		}
		for _, worker := range appSpec.Workers {
			if worker.Name != allFiles[key].Name {
				continue
			}
			worker.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: allFiles[key].Repository, Tag: allFiles[key].Tag}
		}
		for _, job := range appSpec.Jobs {
			if job.Name != allFiles[key].Name {
				continue
			}
			job.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: allFiles[key].Repository, Tag: allFiles[key].Tag}
		}

	}

}

func main() {
	//import and return json object of changed repo
	input, appSpec, err := getAllRepo("test1", "temp.yaml")
	if err != nil {
		fmt.Println("Error in Retrieving json data: ", err)
		os.Exit(1)
	}
	execCommand(input, appSpec)

	newJson, err := json.Marshal(appSpec)
	if err != nil {
		log.Fatal("Error in building json spec")
		os.Exit(1)
	}
	err = ioutil.WriteFile("spec.json", newJson, 0644)
	if err != nil {
		log.Fatal("Error in writing json spec")
		os.Exit(1)
	}

}
