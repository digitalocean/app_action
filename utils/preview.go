package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
)

// SanitizeSpecForPullRequestPreview modifies the given AppSpec to be suitable for a pull request preview.
// This includes:
// - Setting a unique app name.
// - Unsetting any domains.
// - Unsetting any alerts.
// - Setting the reference of all relevant components to point to the PRs ref.
func SanitizeSpecForPullRequestPreview(spec *godo.AppSpec, ghCtx *gha.GitHubContext) error {
	repoOwner, repo := ghCtx.Repo()

	// Override app name to something that identifies this PR.
	spec.Name = GenerateAppName(repoOwner, repo, ghCtx.RefName)

	// Unset any domains as those might collide with production apps.
	spec.Domains = nil

	// Unset any alerts as those will be delivered wrongly anyway.
	spec.Alerts = nil

	// Override the reference of all relevant components to point to the PRs ref.
	if err := godo.ForEachAppSpecComponent(spec, func(c godo.AppBuildableComponentSpec) error {
		// TODO: Should this also deal with raw Git sources?
		ref := c.GetGitHub()
		if ref == nil || ref.Repo != fmt.Sprintf("%s/%s", repoOwner, repo) {
			// Skip Github refs pointing to other repos.
			return nil
		}
		// We manually kick new deployments so we can watch their status better.
		ref.DeployOnPush = false
		ref.Branch = ghCtx.HeadRef
		return nil
	}); err != nil {
		return fmt.Errorf("failed to sanitize buildable components: %w", err)
	}
	return nil
}

// GenerateAppName generates a unique app name based on the repoOwner, repo, and ref.
func GenerateAppName(repoOwner, repo, ref string) string {
	baseName := fmt.Sprintf("%s-%s-%s", repoOwner, repo, ref)
	baseName = strings.ToLower(baseName)
	baseName = strings.NewReplacer(
		"/", "-", // Replace slashes.
		":", "", // Colons are illegal.
		"_", "-", // Underscores are illegal.
	).Replace(baseName)

	// Generate a hash from the unique enumeration of repoOwner, repo, and ref.
	hasher := sha256.New()
	hasher.Write([]byte(baseName))
	suffix := "-" + hex.EncodeToString(hasher.Sum(nil))[:8]

	// App names must be at most 32 characters.
	limit := 32 - len(suffix)
	if len(baseName) < limit {
		limit = len(baseName)
	}

	return baseName[:limit] + suffix
}
