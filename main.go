package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}

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
		// fmt.Println(all_files[key].Name)

		// fmt.Println(all_files[key].Repository)

		// fmt.Println(all_files[key].Tag)

		cmd := exec.Command("sh", "-c", `yq eval '.*[]| select(.name == "`+all_files[key].Name+`").image.repository |=  "`+all_files[key].Repository+
			`" |`+`select(.name == "`+all_files[key].Name+`").image.tag |=  "`+all_files[key].Tag+`"' app.yaml`)
		stdout, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)

		}
		fmt.Print(string(stdout))
	}
	//import and check for the viper yaml
	// vi := viper.New()
	// vi.SetConfigFile("app.yaml")

}
