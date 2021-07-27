package mylib

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

type DoctlDependencies interface {
	RetrieveActiveDeployment(appID string) ([]byte, error)
	UpdateAppPlatformAppSpec(appID string) error
	IsAuthenticated(token string) error
	GetCurrentDeployment(appID string) ([]byte, error)
	CreateDeployments(appID string) error
	RetrieveFromDigitalocean() ([]byte, error)
}
type DoctlServices struct {
	dep DoctlDependencies
}

// UpdatedRepo used for parsing json object of changed repo
type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}

func (d *DoctlServices) IsAuthenticated(token string) error {
	val, err := exec.Command("sh", "-c", fmt.Sprintf("doctl auth init --access-token %s", token)).Output()
	if err != nil {
		return fmt.Errorf("unable to authenticate user: %s", val)
	}
	return nil
}

//getCurrentDeployment returns the current deployment
func (d *DoctlServices) GetCurrentDeployment(appID string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app list-deployments %s -ojson", appID))
	spec, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error in retrieving list of deployments")
	}
	return spec, nil

}

//retrieveActiveDeploymentID retrieves currently deployed app spec of on App Platform app
func (d *DoctlServices) RetrieveActiveDeploymentID(appID string) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl apps get --format ActiveDeployment.ID --no-header %s", appID))
	deployID, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve active deployment")
	}
	deploymentID := strings.TrimSpace(string(deployID))
	return deploymentID, nil

}

//retrieveActiveDeployment returns the active deployment from deplyment appID
func (d *DoctlServices) RetrieveActiveDeployment(deploymentID string, appID string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app get-deployment %s %s -ojson", appID, string(deploymentID)))
	apps, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error in retrieving currently deployed app id")
	}
	return apps, nil
}

//updateAppPlatformAppSpec updates app spec and creates deployment for the app
func (d *DoctlServices) UpdateAppPlatformAppSpec(appID string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app update %s --spec .do._app.yaml", appID))
	_, err := cmd.Output()
	if err != nil {
		fmt.Print(err)
		return errors.Wrap(err, "unable to update app")
	}
	err = d.CreateDeployments(appID)
	if err != nil {
		return err
	}
	return nil
}

//createDeployments creates deployment for the app
func (d *DoctlServices) CreateDeployments(appID string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app create-deployment %s", appID))
	_, err := cmd.Output()
	if err != nil {
		return errors.Wrap(err, "unable to create-deployment for app")
	}
	return nil
}

//retrieveFromDigitalocean returns the app from digitalocean
func (d *DoctlServices) RetrieveFromDigitalocean() ([]byte, error) {
	cmd := exec.Command("sh", "-c", "doctl app list -ojson")
	apps, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user app data from digitalocean")
	}
	return apps, nil

}

// retrieveAppID retrieves app id from app platform
func (d *DoctlServices) RetrieveAppID(appName string) (string, error) {
	apps, err := d.RetrieveFromDigitalocean()
	if err != nil {
		return "", err
	}
	//parsing incoming data for AppId
	var arr []godo.App
	err = json.Unmarshal(apps, &arr)
	if err != nil {
		return "", errors.Wrap(err, "error in parsing data for AppId")
	}
	var appID string

	for k := range arr {
		if arr[k].Spec.Name == appName {
			appID = arr[k].ID
			break
		}
	}
	if appID == "" {
		return "", errors.Wrap(err, "app not found")
	}

	return appID, nil
}

//isDeployed checks for the status of deployment
func (d *DoctlServices) IsDeployed(appID string) error {
	done := false
	for !done {
		fmt.Println("App Platform is Building ....")
		spec, err := d.GetCurrentDeployment(appID)
		if err != nil {
			return errors.Wrap(err, "error in retrieving list of deployments")
		}
		var app []godo.Deployment
		err = json.Unmarshal(spec, &app)
		if err != nil {
			return errors.Wrap(err, "error in parsing deployment")
		}
		if app[0].Phase == "ACTIVE" {
			fmt.Println("Build successful")
			return nil
		}
		if app[0].Phase == "Failed" {
			fmt.Println("Build unsuccessful")
			return errors.Wrap(err, "build unsuccessful")
		}
	}
	return nil
}

// getAllRepo reads the file and return json object of type UpdatedRepo
func (d *DoctlServices) GetAllRepo(input string, appName string) ([]UpdatedRepo, error) {
	//takes care of empty input for Deployment
	if strings.TrimSpace(string(input)) == "" {
		appID, err := d.RetrieveAppID(appName)
		if err != nil {
			return nil, err
		}
		err = d.CreateDeployments(appID)
		if err != nil {
			return nil, err
		}
		error := d.IsDeployed(appID)
		if error != nil {
			return nil, error
		}
		if error == nil {
			os.Exit(0)
		}
		return nil, nil
	}
	var allRepos []UpdatedRepo
	err := json.Unmarshal([]byte(input), &allRepos)
	if err != nil {
		return nil, errors.Wrap(err, "error in parsing json data from file")
	}
	return allRepos, nil
}
