package main

import (
	"encoding/json"
	"errors"
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

type AllError struct {
	name     string
	notFound []string
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

func execCommand(allFiles []UpdatedRepo, appSpec godo.AppSpec) (error, AllError) {
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
				service.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: allFiles[key].Repository, Tag: allFiles[key].Tag}
				delete(nameMap, service.Name)
			}
		}
		for _, worker := range appSpec.Workers {
			if worker.Name != allFiles[key].Name {
				continue
			}
			worker.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: allFiles[key].Repository, Tag: allFiles[key].Tag}
			delete(nameMap, worker.Name)
		}
		for _, job := range appSpec.Jobs {
			if job.Name != allFiles[key].Name {
				continue
			}
			job.Image = &godo.ImageSourceSpec{RegistryType: "DOCR", Repository: allFiles[key].Repository, Tag: allFiles[key].Tag}
			delete(nameMap, job.Name)
		}

	}
	if len(nameMap) == 0 {
		return nil, AllError{}
	} else {
		keys := make([]string, 0, len(nameMap))
		for k := range nameMap {
			keys = append(keys, k)
		}
		error_new := AllError{
			name:     "all files not found",
			notFound: keys,
		}
		return errors.New(error_new.name), error_new
	}

}

func main() {
	//import and return json object of changed repo
	input, appSpec, err := getAllRepo("test1", "app.yaml")
	if err != nil {
		fmt.Println("Error in Retrieving json data: ", err)
		os.Exit(1)
	}
	err, new_err := execCommand(input, appSpec)
	if err != nil {
		fmt.Println(new_err.name)
		fmt.Printf("%v", new_err.notFound)
		os.Exit(1)
	}
	newYaml, err := yaml.Marshal(appSpec)
	if err != nil {
		log.Fatal("Error in building json spec")
		os.Exit(1)
	}
	err = ioutil.WriteFile("app.yaml", newYaml, 0644)
	if err != nil {
		log.Fatal("Error in writing to yaml")
		os.Exit(1)
	}

}
