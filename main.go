package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

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
func getAllRepo(input string) ([]UpdatedRepo, error) {
	//parsing input
	jsonByteValue, err := readFileFrom(input)
	if err != nil {
		log.Fatal("Error in reading from file: ", err)
		return nil, err
	}
	var allRepos []UpdatedRepo
	err = json.Unmarshal(jsonByteValue, &allRepos)
	if err != nil {
		log.Fatal("Error in parsing json data from file: ", err)
		return nil, err
	}
	return allRepos, nil

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

func execCommand(allFiles []UpdatedRepo, appSpec godo.AppSpec) AllError {
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

func uploadToDOCR(data []UpdatedRepo) error {

	for k, _ := range data {
		cmd := exec.Command("sh", "-c", `docker push `+data[k].Repository)
		_, err := cmd.Output()
		if err != nil {
			log.Fatal("Unable to upload image to docr app:", data[k].Name)
			return err
		}
	}
	return nil

}
func main() {
	//import and return json object of changed repo
	//authenticate
	cmd := exec.Command("sh", "-c", "doctl auth init")
	_, err := cmd.Output()
	if err != nil {
		log.Fatal("Unable to retrieve app:", err)
		os.Exit(1)
	}
	//retrieve AppId from users deployment
	cmd = exec.Command("sh", "-c", "doctl app list -ojson")
	apps, err := cmd.Output()
	if err != nil {
		log.Fatal("Unable to retrieve app:", err)
		os.Exit(1)
	}
	//parsing incoming data from AppId
	var arr []godo.App
	err = json.Unmarshal(apps, &arr)
	if err != nil {
		log.Fatal("Error in retrieving app id", err)
		os.Exit(1)
	}
	var appId string
	for k, _ := range arr {
		if arr[k].Spec.Name == "sample-monorepo" {
			appId = arr[k].ID
			break
		}
	}
	if appId == "" {
		log.Fatal("Unable to retrieve appId")
		os.Exit(1)
	}
	//get app based on appID
	cmd = exec.Command("sh", "-c", `doctl app get `+appId+` -ojson`)
	apps, err = cmd.Output()
	if err != nil {
		log.Fatal("Unable to retrieve app:", err)
		os.Exit(1)
	}
	var app []godo.App
	err = json.Unmarshal(apps, &app)
	if err != nil {
		fmt.Println("Error in retrieving app spec: ", err)
		os.Exit(1)
	}
	appSpec := *app[0].Spec
	input, err := getAllRepo("test1")
	if err != nil {
		fmt.Println("Error in Retrieving json data: ", err)
		os.Exit(1)
	}
	//docr registry login
	cmd = exec.Command("sh", "-c", "doctl registry login")
	_, err = cmd.Output()
	if err != nil {
		log.Fatal("Unable to login to digitalocean registry:", err)
		os.Exit(1)
	}
	err = uploadToDOCR(input)
	if err != nil {
		log.Fatal("DOCR update error occured")
		os.Exit(1)
	}
	//updates all the docr images based on users input
	new_err := execCommand(input, appSpec)
	if new_err.name != "" {
		fmt.Println(new_err.name)
		fmt.Printf("%v", new_err.notFound)
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

}
