package parser_struct

import "github.com/digitalocean/godo"

// UpdatedRepo used for parsing json object of changed repo
type UpdatedRepo struct {
	// Name is the App Component Name.
	Name string `json:"name,omitempty"`
	// Repo is the Repository to be deployed.
	// Deprecated: Use Image instead.
	Repository string `json:"repository,omitempty"`
	// Tag is the image tag to be deployed.
	// Deprecated: Use Image instead.
	Tag string `json:"tag,omitempty"`
	// Image is the ImageSourceSpec to apply to the component.
	Image godo.ImageSourceSpec `json:"image,omitempty"`
}
