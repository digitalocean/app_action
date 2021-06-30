package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/digitalocean/godo"
	"sigs.k8s.io/yaml"
)

//used for parsing json object of changed repo
type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
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

//parsing app
func parseAppSpec(spec []byte) (*godo.AppSpec, error) {
	jsonSpec, err := yaml.YAMLToJSON(spec)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(jsonSpec))
	dec.DisallowUnknownFields()

	var appSpec godo.AppSpec
	if err := dec.Decode(&appSpec); err != nil {
		return nil, err
	}

	return &appSpec, nil
}

//reads the file and return json object of type UpdatedRepo
func getAllRepo(input string, appSpec string) ([]UpdatedRepo, godo.AppSpec, error) {
	//parsing input
	jsonByteValue, err := readFileFrom(input)
	if err != nil {
		log.Fatal("Error in reading from file: ", err)
		return nil, godo.AppSpec{}, err
	}
	var allRepos []UpdatedRepo
	err = json.Unmarshal(jsonByteValue, &allRepos)
	if err != nil {
		log.Fatal("Error in parsing json data from file: ", err)
		return nil, godo.AppSpec{}, err
	}
	//parsing yml
	yamlByteValue, err := readFileFrom(appSpec)
	if err != nil {
		log.Fatal("Error in reading from file: ", err)
		return nil, godo.AppSpec{}, err
	}
	var spec godo.AppSpec
	err = yaml.Unmarshal(yamlByteValue, &spec)
	if err != nil {
		return nil, godo.AppSpec{}, err
	}
	fmt.Printf("%+v\n", spec)
	return allRepos, spec, nil

}
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

func execCommand(allFiles []UpdatedRepo, appSpec godo.AppSpec) {
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
