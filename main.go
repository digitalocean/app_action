package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/viper"
)

type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}

func getAllRepo(fileLocation string) ([]UpdatedRepo, error) {
	jsonFile, err := os.Open(fileLocation)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var allRepos []UpdatedRepo
	json.Unmarshal(byteValue, &allRepos)
	return allRepos, nil

}

func main() {
	//import and return json object of changed repo
	_, err := getAllRepo("../test1.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//import and check for the viper yaml
	vi := viper.New()
	vi.SetConfigFile("app.yaml")
	vi.SetConfigType("yaml")
	err = vi.ReadInConfig()
	if err != nil {
		fmt.Println("Error in app spec retrieved from DO ", err)
		os.Exit(1)
	}
	fmt.Println(vi.GetString("static_sites.fronted"))

}
