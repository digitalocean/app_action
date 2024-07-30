package main

import (
	"context"
	"net/http"

	"github.com/digitalocean/app_action/utils"
	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
)

func main() {
	ctx := context.Background()
	a := gha.New()

	in, err := getInputs(a)
	if err != nil {
		a.Fatalf("failed to get inputs: %v", err)
	}
	// Mask the DO token to avoid accidentally leaking it.
	a.AddMask(in.token)

	if in.appID == "" && in.appName == "" && !in.fromPRPreview {
		a.Fatalf("either app_id, app_name, or from_pr_preview must be set")
	}

	ghCtx, err := a.Context()
	if err != nil {
		a.Fatalf("failed to get GitHub context: %v", err)
	}

	do := godo.NewFromToken(in.token).Apps

	appID := in.appID
	if appID == "" {
		appName := in.appName
		if appName == "" {
			repoOwner, repo := ghCtx.Repo()
			appName = utils.GenerateAppName(repoOwner, repo, ghCtx.RefName)
		}

		app, err := utils.FindAppByName(ctx, do, appName)
		if err != nil {
			a.Fatalf("failed to find app: %v", err)
		}
		if app == nil {
			if in.ignoreNotFound {
				a.Infof("app %q not found, ignoring", appName)
				return
			}
			a.Fatalf("app %q not found", appName)
		}
		appID = app.ID
	}

	if resp, err := do.Delete(ctx, appID); err != nil {
		if resp.StatusCode == http.StatusNotFound && in.ignoreNotFound {
			a.Infof("app %q not found, ignoring", appID)
			return
		}
		a.Fatalf("failed to delete app: %v", err)
	}
}
