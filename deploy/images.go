package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/digitalocean/godo"
)

// replaceImagesInSpec replaces the images in the given AppSpec with the ones defined in the environment.
func replaceImagesInSpec(spec *godo.AppSpec) error {
	if err := godo.ForEachAppSpecComponent(spec, func(c godo.AppContainerComponentSpec) error {
		image := c.GetImage()
		if image == nil {
			return nil
		}

		if digest := os.Getenv("IMAGE_DIGEST_" + componentNameToEnvVar(c.GetName())); digest != "" {
			image.Tag = ""
			image.Digest = digest
		} else if tag := os.Getenv("IMAGE_TAG_" + componentNameToEnvVar(c.GetName())); tag != "" {
			image.Digest = ""
			image.Tag = tag
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to sanitize buildable components: %w", err)
	}
	return nil
}

// componentNameToEnvVar converts a component name to an environment variable name.
func componentNameToEnvVar(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}
