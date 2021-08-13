package doctl

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/digitalocean/app_action/internal/parser"
	"github.com/digitalocean/app_action/internal/parser_struct"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

// Client is a struct for holding doctl dependent function interface
type Client struct {
}

// NewClient doctl client wrapper
func NewClient(token string) (Client, error) {
	val, err := exec.Command("sh", "-c", fmt.Sprintf("doctl auth init --access-token %s", token)).Output()
	if err != nil {
		return Client{}, fmt.Errorf("unable to authenticate user: %s", val)
	}

	d := Client{}

	return d, nil
}

// ListDeployments takes appID as input and returns list of deployments (used to retrieve most recent deployment)
func (d *Client) ListDeployments(appID string) ([]godo.Deployment, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app list-deployments %s -ojson", appID))
	spec, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error in retrieving list of deployments")
	}

	// parsing incoming data to get all deployments
	deployments, err := parser.ParseDeploymentSpec(spec)
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

// RetrieveActiveDeploymentID takes appID as input and retrieves currently deployment id of the active deployment of the app on App Platform
func (d *Client) RetrieveActiveDeploymentID(appID string) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl apps get --format ActiveDeployment.ID --no-header %s", appID))
	deployID, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve active deployment")
	}
	deploymentID := strings.TrimSpace(string(deployID))
	return deploymentID, nil
}

// RetrieveActiveDeployment takes active deployment id as input from(RetrieveActiveDeploymentID) and app id
// returns the app spec from App Platform as *godo.AppSpec, retrieves parsed json object of the json input
func (d *Client) RetrieveActiveDeployment(deploymentID string, appID string, input string) ([]parser_struct.UpdatedRepo, *godo.AppSpec, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app get-deployment %s %s -ojson", appID, string(deploymentID)))
	apps, err := cmd.Output()
	if err != nil {
		return nil, nil, errors.Wrap(err, "error in retrieving currently deployed app id")
	}

	//parse json input
	allRepos, err := parser.ParseJsonInput(input)
	if err != nil {
		return nil, nil, err
	}

	//parse deployment spec
	deployments, err := parser.ParseDeploymentSpec(apps)
	if err != nil {
		return nil, nil, err
	}
	return allRepos, deployments[0].Spec, nil
}

// UpdateAppPlatformAppSpec takes appID as input
// updates App Platform's app spec and creates deployment
func (d *Client) UpdateAppPlatformAppSpec(tmpfile, appID string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app update %s --spec %s", appID, tmpfile))
	_, err := cmd.Output()
	if err != nil {
		fmt.Printf("doctl app update %s --spec %s", appID, tmpfile)
		return errors.Wrap(err, "unable to update app")
	}
	return nil
}

// CreateDeployments takes app id as an input and creates deployment for the app
func (d *Client) CreateDeployments(appID string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("doctl app create-deployment %s", appID))
	_, err := cmd.Output()
	if err != nil {
		return errors.Wrap(err, "unable to create-deployment for app")
	}
	return nil
}

// RetrieveFromDigitalocean returns the app from DigitalOcean as a slice of byte
func (d *Client) RetrieveFromDigitalocean() ([]godo.App, error) {
	cmd := exec.Command("sh", "-c", "doctl app list -ojson")
	apps, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user app data from digitalocean")
	}
	// parsing incoming data for AppId
	arr, err := parser.ParseAppSpec(apps)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// RetrieveAppID takes unique app name as an input and retrieves app id from app platform based on the users unique app name
func (d *Client) RetrieveAppID(appName string) (string, error) {
	arr, err := d.RetrieveFromDigitalocean()
	if err != nil {
		return "", err
	}
	//retrieve app id app array
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

// IsDeployed takes app id as an input and checks for the status of the deployment until the status is updated to ACTIVE or failed
func (d *Client) IsDeployed(appID string) error {
	done := false
	fmt.Println("App Platform is Building ....")
	for !done {
		app, err := d.ListDeployments(appID)
		if err != nil {
			return errors.Wrap(err, "error in retrieving list of deployments")
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

// Deploy redeploys app if user provides empty json file
func (d *Client) Deploy(input string, appName string) error {
	if strings.TrimSpace(string(input)) == "" {
		appID, err := d.RetrieveAppID(appName)
		if err != nil {
			return err
		}
		err = d.CreateDeployments(appID)
		if err != nil {
			return err
		}
		err = d.IsDeployed(appID)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.Errorf("Please provide valid json input")
}
