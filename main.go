package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/digitalocean/godo"
	"gopkg.in/yaml.v2"
)

//used for parsing json object of changed repo
type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}

//reads the file from fileLocation
func readFileFrom(fileLocation string) ([]byte, error) {
	jsonFile, err := os.Open(fileLocation)
	if err != nil {
		log.Fatal("Error in opening the file", err)
		return []byte{}, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
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

func execCommand(allFiles []UpdatedRepo, appSpec godo.AppSpec) error {
	checkForGitAndDockerHub(allFiles, &appSpec)
	for key, _ := range allFiles {
		cmd := exec.Command("sh", "-c", `cat _temp.yaml |yq eval '(.*[]| select(.name == "`+allFiles[key].Name+`").image.repository) |=  "`+allFiles[key].Repository+
			`" |`+`(.*[]| select(.name == "`+allFiles[key].Name+`").image.registry_type) |= "DOCR" |(.*[]|select(.name == "`+allFiles[key].Name+`").image.tag) |=  "`+allFiles[key].Tag+`"' -| sponge _temp.yaml`)
		_, err := cmd.Output()
		if err != nil {
			log.Fatal("Error in checking docr path file: ", err)
			os.Exit(1)
		}

	}
	return nil

}

func main() {
	//import and return json object of changed repo
	input, appSpec, err := getAllRepo("test1", "_temp.yaml")
	if err != nil {
		fmt.Println("Error in Retrieving json data: ", err)
		os.Exit(1)
	}
	err = execCommand(input, appSpec)
	if err != nil {
		log.Fatal("Error in retrieving data from")
	}

}
