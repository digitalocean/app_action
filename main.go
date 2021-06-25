package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
		log.Fatal(err)
		os.Exit(1)
		return []byte{}, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
		return []byte{}, err
	}
	return byteValue, err
}

//reads the file and return json object of type UpdatedRepo
func getAllRepo(location string) ([]UpdatedRepo, error) {
	byteValue, err := readFileFrom(location)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
		return nil, err
	}
	var allRepos []UpdatedRepo
	json.Unmarshal(byteValue, &allRepos)
	return allRepos, nil

}

func main() {
	//import and return json object of changed repo
	all_files, err := getAllRepo("../test1.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for key, _ := range all_files {

		cmd := exec.Command("sh", "-c", `sudo cat temp.yaml |yq eval '(.*[]| select(.name == "`+all_files[key].Name+`").image.repository) |=  "`+all_files[key].Repository+
			`" |`+`(.*[]|select(.name == "`+all_files[key].Name+`").image.tag) |=  "`+all_files[key].Tag+`"' -| sudo sponge temp.yaml`)
		_, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)

		}

	}

}
