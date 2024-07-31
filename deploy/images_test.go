package main

import (
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/require"
)

func TestReplaceImagesInSpec(t *testing.T) {
	spec := &godo.AppSpec{
		Name: "foo",
		Services: []*godo.AppServiceSpec{{
			Name: "web",
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_Ghcr,
				Registry:     "foo",
				Repository:   "bar",
				Tag:          "latest",
			},
		}},
		Workers: []*godo.AppWorkerSpec{{
			Name: "fancy-worker",
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_DockerHub,
				Registry:     "foo",
				Repository:   "worker",
				Tag:          "latest",
			},
		}},
		Jobs: []*godo.AppJobSpec{{
			Name: "job",
			GitHub: &godo.GitHubSourceSpec{
				Repo:   "foo/bar",
				Branch: "main",
			},
		}},
	}

	t.Setenv("IMAGE_TAG_WEB", "v1")
	t.Setenv("IMAGE_DIGEST_FANCY_WORKER", "1234abcd")
	t.Setenv("IMAGE_DIGEST_JOB", "1234abcd")
	err := replaceImagesInSpec(spec)
	require.NoError(t, err)

	expected := &godo.AppSpec{
		Name: "foo",
		Services: []*godo.AppServiceSpec{{
			Name: "web",
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_Ghcr,
				Registry:     "foo",
				Repository:   "bar",
				Tag:          "v1", // Tag was updated.
			},
		}},
		Workers: []*godo.AppWorkerSpec{{
			Name: "fancy-worker",
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_DockerHub,
				Registry:     "foo",
				Repository:   "worker",
				Digest:       "1234abcd", // Digest was updated, tag was removed.
			},
		}},
		Jobs: []*godo.AppJobSpec{{
			Name: "job",
			GitHub: &godo.GitHubSourceSpec{
				Repo:   "foo/bar", // No change.
				Branch: "main",
			},
		}},
	}

	require.Equal(t, expected, spec)
}
