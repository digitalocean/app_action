package utils

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// FindAppByName returns the app with the given name, or nil if it does not exist.
func FindAppByName(ctx context.Context, ap godo.AppsService, name string) (*godo.App, error) {
	opt := &godo.ListOptions{}
	for {
		apps, resp, err := ap.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list apps: %w", err)
		}

		for _, a := range apps {
			if a.GetSpec().GetName() == name {
				return a, nil
			}
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, fmt.Errorf("failed to get current page: %w", err)
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}
	return nil, nil
}
