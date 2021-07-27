package mylib

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

//DoctlDependencies interface for doctl dependent functions
type DoctlDependencies interface {
	IsAuthenticated(token string) error
	GetCurrentDeployment(appID string) ([]byte, error)
	RetrieveActiveDeploymentID(appID string) (string, error)
	RetrieveActiveDeployment(deploymentID string, appID string) ([]byte, error)
	UpdateAppPlatformAppSpec(appID string) error
	CreateDeployments(appID string) error
	RetrieveFromDigitalocean() ([]byte, error)
	RetrieveAppID(appName string) (string, error)
	IsDeployed(appID string) error
	ReDeploy(input string, appName string) error
}

//DoctlServices is a struct for holding doctl dependent functions
type DoctlServices struct {
	Dep DoctlDependencies
}

//UpdatedRepo used for parsing json object of changed repo
type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}

//IsAuthenticated used for user authentication
func (d *DoctlServices) IsAuthenticated(token string) error {
	val, err := exec.Command("sh", "-c", fmt.Sprintf("doctl auth init --access-token %s", token)).Output()
	if err != nil {
		return fmt.Errorf("unable to authenticate user: %s", val)
	}
	return nil
}

//GetCurrentDeployment takes appID as input and returns list of deployments (used to retrieve most recent deployment)
func (d *DoctlServices) GetCurrentDeployment(appID string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app list-deployments %s -ojson", appID))
	spec, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error in retrieving list of deployments")
	}
	return spec, nil

}

//RetrieveActiveDeploymentID takes appID as input and retrieves currently deploymentID of the active deployment of the app on App Platform
func (d *DoctlServices) RetrieveActiveDeploymentID(appID string) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl apps get --format ActiveDeployment.ID --no-header %s", appID))
	deployID, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve active deployment")
	}
	deploymentID := strings.TrimSpace(string(deployID))
	return deploymentID, nil

}

//RetrieveActiveDeployment takes active deployment id as input from(RetrieveActiveDeploymentID) and appID
//returns the app spec from App Platform as []byte
func (d *DoctlServices) RetrieveActiveDeployment(deploymentID string, appID string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app get-deployment %s %s -ojson", appID, string(deploymentID)))
	apps, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error in retrieving currently deployed app id")
	}
	return apps, nil
}

//UpdateAppPlatformAppSpec takes appID as input
//This function updates App Platform's app spec and creates Deployment
func (d *DoctlServices) UpdateAppPlatformAppSpec(appID string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app update %s --spec .do._app.yaml", appID))
	_, err := cmd.Output()
	if err != nil {
		fmt.Print(err)
		return errors.Wrap(err, "unable to update app")
	}
	return nil
}

//CreateDeployments takes appID as an input and creates deployment for the app
func (d *DoctlServices) CreateDeployments(appID string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app create-deployment %s", appID))
	_, err := cmd.Output()
	if err != nil {
		return errors.Wrap(err, "unable to create-deployment for app")
	}
	return nil
}

//RetrieveFromDigitalocean returns the app from DigitalOcean as a slice of byte
func (d *DoctlServices) RetrieveFromDigitalocean() ([]byte, error) {
	cmd := exec.Command("sh", "-c", "doctl app list -ojson")
	apps, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user app data from digitalocean")
	}
	return apps, nil

}

// RetrieveAppID takes unique appName as an input and retrieves app id from app platform based on the users unique app name
func (d *DoctlServices) RetrieveAppID(appName string) (string, error) {
	apps, err := d.dep.RetrieveFromDigitalocean()
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

//isDeployed takes appID as an input and checks for the status of the deployment
//until the status is updated to Active deployment or failed
func (d *DoctlServices) IsDeployed(appID string) error {
	done := false
	for !done {
		fmt.Println("App Platform is Building ....")
		spec, err := d.dep.GetCurrentDeployment(appID)
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

//ReDeploy This function then checks if the input is empty and if it is then deploys the app using the existing appSpec
func (d *DoctlServices) ReDeploy(input string, appName string) error {
	if strings.TrimSpace(string(input)) == "" {
		appID, err := d.dep.RetrieveAppID(appName)
		if err != nil {
			return err
		}
		err = d.dep.CreateDeployments(appID)
		if err != nil {
			return err
		}
		err = d.dep.IsDeployed(appID)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.Errorf("Please provide valid json input")

}
