package utils

import (
	"testing"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/require"
)

func TestSanitizeSpecForPullRequestPreview(t *testing.T) {
	spec := &godo.AppSpec{
		Name:    "foo",
		Domains: []*godo.AppDomainSpec{{Domain: "foo.com"}},
		Alerts:  []*godo.AppAlertSpec{{Value: 80}},
		Services: []*godo.AppServiceSpec{{
			Name: "web",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "foo/bar",
				Branch:       "main",
				DeployOnPush: true,
			},
		}, {
			Name: "web2",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "another/repo",
				Branch:       "main",
				DeployOnPush: true,
			},
		}},
		Workers: []*godo.AppWorkerSpec{{
			Name: "worker",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "foo/bar",
				Branch:       "main",
				DeployOnPush: true,
			},
		}},
		Jobs: []*godo.AppJobSpec{{
			Name: "job",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "foo/bar",
				Branch:       "main",
				DeployOnPush: true,
			},
		}},
		Functions: []*godo.AppFunctionsSpec{{
			Name: "function",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "foo/bar",
				Branch:       "main",
				DeployOnPush: true,
			},
		}},
	}

	ghCtx := &gha.GitHubContext{
		Repository: "foo/bar",
		RefName:    "3/merge",
		HeadRef:    "feature-branch",
	}

	err := SanitizeSpecForPullRequestPreview(spec, ghCtx)
	require.NoError(t, err)

	expected := &godo.AppSpec{
		Name: "foo-bar-3-merge-adb46530", // Name got generated.
		// Domains and alerts got removed.
		Services: []*godo.AppServiceSpec{{
			Name: "web",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "foo/bar",
				Branch:       "feature-branch", // Branch got updated.
				DeployOnPush: false,            // DeployOnPush got set to false.
			},
		}, {
			Name: "web2",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "another/repo", // No change.
				Branch:       "main",
				DeployOnPush: true,
			},
		}},
		Workers: []*godo.AppWorkerSpec{{
			Name: "worker",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "foo/bar",
				Branch:       "feature-branch", // Branch got updated.
				DeployOnPush: false,            // DeployOnPush got set to false.
			},
		}},
		Jobs: []*godo.AppJobSpec{{
			Name: "job",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "foo/bar",
				Branch:       "feature-branch", // Branch got updated.
				DeployOnPush: false,            // DeployOnPush got set to false.
			},
		}},
		Functions: []*godo.AppFunctionsSpec{{
			Name: "function",
			GitHub: &godo.GitHubSourceSpec{
				Repo:         "foo/bar",
				Branch:       "feature-branch", // Branch got updated.
				DeployOnPush: false,            // DeployOnPush got set to false.
			},
		}},
	}

	require.Equal(t, expected, spec)
}

func TestGenerateAppName(t *testing.T) {
	tests := []struct {
		name      string
		repoOwner string
		repo      string
		ref       string
		expected  string
	}{{
		name:      "success",
		repoOwner: "foo",
		repo:      "bar",
		ref:       "3/merge",
		expected:  "foo-bar-3-merge-adb46530",
	}, {
		name:      "long repo owner",
		repoOwner: "thisisanextremelylongrepohostname",
		repo:      "bar",
		ref:       "3/merge",
		expected:  "thisisanextremelylongre-92da974b",
	}, {
		name:      "long repo",
		repoOwner: "foo",
		repo:      "thisisanextremelylongreponame",
		ref:       "3/merge",
		expected:  "foo-thisisanextremelylo-67dbc40d",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := GenerateAppName(test.repoOwner, test.repo, test.ref)
			require.Equal(t, test.expected, got)
		})
	}
}
