package main

import (
	"os"
	"reflect"
	"testing"
)

func testingFileCreation() (string, error) {
	os.Remove("_test")
	testFileInput := `[
	{
	  "name": "frontend",
	  "repository": "registry.digitalocean.com/<my-registry>/<my-image>",
	  "tag": "latest"
	},
	{
	  "name": "landing",
	  "repository": "registry.digitalocean.com/<my-registry>/<my-image>",
	  "tag": "test1"
	},
	{
	  "name": "api",
	  "repository": "registry.digitalocean.com/<my-registry>/<my-image>",
	  "tag": "test2"
	}
  ]`
	testFile := []byte(testFileInput)
	file, err := os.Create("_test")
	if err != nil {
		return "", err
	}
	_, err = file.Write(testFile)
	if err != nil {
		return "", err
	}
	defer(os.Remove("_test"))
	return testFileInput, nil
}
func TestReadFileFrom(t *testing.T) {
	//Test to check if read is working correctly
	//For this I will read test1 file and verify the output

	testFileInput, err := testingFileCreation()
	if err != nil {
		t.Error("Error in file Creation: ", err)
	}
	jsonFile, err := readFileFrom("_test")
	if err != nil {
		t.Error("Unable to read file", err)
	}
	if string(jsonFile) != testFileInput {
		t.Error("mismatched file: ", testFileInput)
	}

}

func TestGetAllRepo(t *testing.T) {
	_, err := testingFileCreation()
	if err != nil {
		t.Error("Error in file Creation: ", err)
	}
	allRepo,spec,err := getAllRepo("_test","_temp")
	if err != nil {
		t.Error("Error in parsing json data")
	}
	var temp = []UpdatedRepo{
		{
			"frontend",
			"registry.digitalocean.com/<my-registry>/<my-image>",
			"latest",
		},
		{
			"landing",
			"registry.digitalocean.com/<my-registry>/<my-image>",
			"test1",
		},
		{
			"api",
			"registry.digitalocean.com/<my-registry>/<my-image>",
			"test2",
		},
	}
	if !reflect.DeepEqual(allRepo, temp) {
		t.Errorf("Error in retrieving struct from json")
	}

	os.Remove("_test")

}
func TestExecCommand(t *testing.T) {

}
