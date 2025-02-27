package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestCreateSpecFromFile(t *testing.T) {
	spec := &godo.AppSpec{
		Name: "foo",
		Envs: []*godo.AppVariableDefinition{{
			Key:   "GLOBAL_ENV_VAR",
			Value: "${APP_DOMAIN}",
		}},
		Services: []*godo.AppServiceSpec{{
			Name: "web",
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_Ghcr,
				Registry:     "foo",
				Repository:   "bar",
				Tag:          "${ENV_VAR}",
			},
			Envs: []*godo.AppVariableDefinition{{
				Key:   "SERVICE_ENV_VAR",
				Value: "${web2.HOSTNAME}",
			}},
		}, {
			Name: "web2",
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_Ghcr,
				Registry:     "foo",
				Repository:   "bar",
				Tag:          "latest",
			},
		}},
	}

	bs, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("failed to marshal spec: %v", err)
	}
	specFilePath := t.TempDir() + "/spec.yaml"
	if err := os.WriteFile(specFilePath, bs, 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	d := &deployer{
		inputs: inputs{appSpecLocation: specFilePath},
	}

	t.Setenv("ENV_VAR", "v1")        // Put in via env substitution.
	t.Setenv("IMAGE_TAG_WEB2", "v2") // Put in via "magic" env var.
	got, err := d.createSpec(context.Background())
	if err != nil {
		t.Fatalf("failed to create spec: %v", err)
	}

	expected := &godo.AppSpec{
		Name: "foo",
		Envs: []*godo.AppVariableDefinition{{
			Key:   "GLOBAL_ENV_VAR",
			Value: "${APP_DOMAIN}", // Bindable reference stayed intact.
		}},
		Services: []*godo.AppServiceSpec{{
			Name: "web",
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_Ghcr,
				Registry:     "foo",
				Repository:   "bar",
				Tag:          "v1", // Tag was updated.
			},
			Envs: []*godo.AppVariableDefinition{{
				Key:   "SERVICE_ENV_VAR",
				Value: "${web2.HOSTNAME}", // Bindable reference stayed intact.
			}},
		}, {
			Name: "web2",
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_Ghcr,
				Registry:     "foo",
				Repository:   "bar",
				Tag:          "v2", // Tag was updated.
			},
		}},
	}

	require.Equal(t, expected, got)
}

func TestCreateSpecFromExistingApp(t *testing.T) {
	tests := []struct {
		name       string
		appService *mockedAppsService
		envs       map[string]string
		expected   *godo.AppSpec
		err        bool
	}{{
		name: "existing app",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", mock.Anything, mock.Anything).Return([]*godo.App{{
				Spec: &godo.AppSpec{
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
				},
			}}, &godo.Response{}, nil)
			return as
		}(),
		envs: map[string]string{"IMAGE_TAG_WEB": "v1"},
		expected: &godo.AppSpec{
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
		},
	}, {
		name: "no app",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			return as
		}(),
		err: true,
	}, {
		name: "error listing apps",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, &godo.Response{}, errors.New("an error"))
			return as
		}(),
		err: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := &deployer{
				apps:   test.appService,
				inputs: inputs{appName: "foo"},
			}

			for k, v := range test.envs {
				t.Setenv(k, v)
			}

			spec, err := d.createSpec(context.Background())
			if err != nil && !test.err {
				require.NoError(t, err)
			}
			if err == nil && test.err {
				require.Error(t, err)
			}
			require.Equal(t, test.expected, spec)
		})
	}
}

func TestDeploy(t *testing.T) {
	ctx := context.Background()
	appID := "app-id"
	deploymentID := "deployment-id"
	spec := &godo.AppSpec{
		Name: "foo",
	}

	tests := []struct {
		name           string
		appService     *mockedAppsService
		logsRT         *mockedRoundtripper
		inputs         inputs
		expectedLogs   []byte
		expectedOutput []byte
		err            bool
	}{{
		name: "success",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			as.On("Create", ctx, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{{
				ID: deploymentID,
			}}, &godo.Response{}, nil)
			as.On("GetDeployment", ctx, appID, deploymentID).Return(&godo.Deployment{
				Phase: godo.DeploymentPhase_Active,
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeBuild, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://build.com"},
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeDeploy, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://deploy.com"},
			}, &godo.Response{}, nil)
			as.On("Get", ctx, appID).Return(&godo.App{ID: appID, LiveURL: "https://example.com"}, &godo.Response{}, nil)
			return as
		}(),
		logsRT: func() *mockedRoundtripper {
			rt := &mockedRoundtripper{}
			rt.On("RoundTrip", mock.Anything).Return(&http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte("build log"))),
			}, nil).Once()
			rt.On("RoundTrip", mock.Anything).Return(&http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte("deploy log"))),
			}, nil).Once()
			return rt
		}(),
		inputs: inputs{
			printBuildLogs:  true,
			printDeployLogs: true,
		},
		expectedLogs: []byte(`app "foo" does not exist yet, creating...
wait for deployment to finish
deployment is in phase: ACTIVE
::group::build logs
build log
::endgroup::
::group::deploy logs
deploy log
::endgroup::
`),
		expectedOutput: []byte(`build_logs<<_GitHubActionsFileCommandDelimeter_
build log
_GitHubActionsFileCommandDelimeter_
deploy_logs<<_GitHubActionsFileCommandDelimeter_
deploy log
_GitHubActionsFileCommandDelimeter_
`),
	}, {
		name: "success on preexisting app",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{{ID: appID, Spec: spec}}, &godo.Response{}, nil)
			as.On("Update", ctx, appID, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{{
				ID: deploymentID,
			}}, &godo.Response{}, nil)
			as.On("GetDeployment", ctx, appID, deploymentID).Return(&godo.Deployment{
				Phase: godo.DeploymentPhase_Active,
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeBuild, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://build.com"},
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeDeploy, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://deploy.com"},
			}, &godo.Response{}, nil)
			as.On("Get", ctx, appID).Return(&godo.App{ID: appID, LiveURL: "https://example.com"}, &godo.Response{}, nil)
			return as
		}(),
		logsRT: func() *mockedRoundtripper {
			rt := &mockedRoundtripper{}
			rt.On("RoundTrip", mock.Anything).Return(&http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte("build log"))),
			}, nil).Once()
			rt.On("RoundTrip", mock.Anything).Return(&http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte("deploy log"))),
			}, nil).Once()
			return rt
		}(),
		expectedLogs: []byte(`app "foo" already exists, updating...
wait for deployment to finish
deployment is in phase: ACTIVE
`),
		expectedOutput: []byte(`build_logs<<_GitHubActionsFileCommandDelimeter_
build log
_GitHubActionsFileCommandDelimeter_
deploy_logs<<_GitHubActionsFileCommandDelimeter_
deploy log
_GitHubActionsFileCommandDelimeter_
`),
	}, {
		name: "fails to deploy",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{{ID: appID, Spec: spec}}, &godo.Response{}, nil)
			as.On("Update", ctx, appID, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{{
				ID: deploymentID,
			}}, &godo.Response{}, nil)
			as.On("GetDeployment", ctx, appID, deploymentID).Return(&godo.Deployment{
				Phase: godo.DeploymentPhase_Error,
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeBuild, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://build.com"},
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeDeploy, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://deploy.com"},
			}, &godo.Response{}, nil)
			as.On("Get", ctx, appID).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			return as
		}(),
		logsRT: func() *mockedRoundtripper {
			rt := &mockedRoundtripper{}
			rt.On("RoundTrip", mock.Anything).Return(&http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte("build log"))),
			}, nil).Once()
			rt.On("RoundTrip", mock.Anything).Return(&http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte("deploy log"))),
			}, nil).Once()
			return rt
		}(),
		err: true,
		expectedLogs: []byte(`app "foo" already exists, updating...
wait for deployment to finish
deployment is in phase: ERROR
`),
		expectedOutput: []byte(`build_logs<<_GitHubActionsFileCommandDelimeter_
build log
_GitHubActionsFileCommandDelimeter_
deploy_logs<<_GitHubActionsFileCommandDelimeter_
deploy log
_GitHubActionsFileCommandDelimeter_
`),
	}, {
		name: "fails to list apps",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, errors.New("an error"))
			return as
		}(),
		err: true,
	}, {
		name: "fails to create app",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			as.On("Create", ctx, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, errors.New("an error"))
			return as
		}(),
		err: true,
		expectedLogs: []byte(`app "foo" does not exist yet, creating...
`),
	}, {
		name: "fails to list deployments",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			as.On("Create", ctx, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{}, &godo.Response{}, errors.New("an error"))
			return as
		}(),
		err: true,
		expectedLogs: []byte(`app "foo" does not exist yet, creating...
`),
	}, {
		name: "returns an empty deployment list",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			as.On("Create", ctx, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{}, &godo.Response{}, nil)
			return as
		}(),
		err: true,
		expectedLogs: []byte(`app "foo" does not exist yet, creating...
`),
	}, {
		name: "fails to get deployment for phase poll",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			as.On("Create", ctx, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{{
				ID: deploymentID,
			}}, &godo.Response{}, nil)
			as.On("GetDeployment", ctx, appID, deploymentID).Return(&godo.Deployment{}, &godo.Response{}, errors.New("an error"))
			return as
		}(),
		err: true,
		expectedLogs: []byte(`app "foo" does not exist yet, creating...
wait for deployment to finish
`),
	}, {
		name: "fails to get get logs",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			as.On("Create", ctx, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{{
				ID: deploymentID,
			}}, &godo.Response{}, nil)
			as.On("GetDeployment", ctx, appID, deploymentID).Return(&godo.Deployment{
				Phase: godo.DeploymentPhase_Active,
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeBuild, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://build.com"},
			}, &godo.Response{Response: &http.Response{StatusCode: http.StatusBadGateway}}, errors.New("an error"))
			return as
		}(),
		err: true,
		expectedLogs: []byte(`app "foo" does not exist yet, creating...
wait for deployment to finish
deployment is in phase: ACTIVE
`),
	}, {
		name: "ignores log failures for 400 returns",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			as.On("Create", ctx, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{{
				ID: deploymentID,
			}}, &godo.Response{}, nil)
			as.On("GetDeployment", ctx, appID, deploymentID).Return(&godo.Deployment{
				Phase: godo.DeploymentPhase_Active,
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeBuild, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://build.com"},
			}, &godo.Response{Response: &http.Response{StatusCode: http.StatusBadRequest}}, errors.New("an error"))
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeDeploy, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://deploy.com"},
			}, &godo.Response{Response: &http.Response{StatusCode: http.StatusBadRequest}}, errors.New("an error"))
			as.On("Get", ctx, appID).Return(&godo.App{ID: appID, LiveURL: "https://example.com"}, &godo.Response{}, nil)
			return as
		}(),
		expectedLogs: []byte(`app "foo" does not exist yet, creating...
wait for deployment to finish
deployment is in phase: ACTIVE
`),
	}, {
		name: "fails to get app for live URL poll",
		appService: func() *mockedAppsService {
			as := &mockedAppsService{}
			as.On("List", ctx, mock.Anything).Return([]*godo.App{}, &godo.Response{}, nil)
			as.On("Create", ctx, mock.Anything).Return(&godo.App{ID: appID}, &godo.Response{}, nil)
			as.On("ListDeployments", ctx, appID, mock.Anything).Return([]*godo.Deployment{{
				ID: deploymentID,
			}}, &godo.Response{}, nil)
			as.On("GetDeployment", ctx, appID, deploymentID).Return(&godo.Deployment{
				Phase: godo.DeploymentPhase_Active,
			}, &godo.Response{}, nil)
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeBuild, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://build.com"},
			}, &godo.Response{Response: &http.Response{StatusCode: http.StatusBadRequest}}, errors.New("an error"))
			as.On("GetLogs", ctx, appID, deploymentID, "", godo.AppLogTypeDeploy, true, -1).Return(&godo.AppLogs{
				HistoricURLs: []string{"http://deploy.com"},
			}, &godo.Response{Response: &http.Response{StatusCode: http.StatusBadRequest}}, errors.New("an error"))
			as.On("Get", ctx, appID).Return(&godo.App{ID: appID, LiveURL: "https://example.com"}, &godo.Response{}, errors.New("an error"))
			return as
		}(),
		err: true,
		expectedLogs: []byte(`app "foo" does not exist yet, creating...
wait for deployment to finish
deployment is in phase: ACTIVE
`),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actionLogs bytes.Buffer
			outputFilePath := t.TempDir() + "/output"
			d := &deployer{
				action: gha.New(gha.WithWriter(&actionLogs), gha.WithGetenv(func(k string) string {
					switch k {
					case "GITHUB_OUTPUT":
						return outputFilePath
					default:
						return ""
					}
				})),
				apps:       test.appService,
				httpClient: &http.Client{Transport: test.logsRT},
				inputs:     test.inputs,
			}
			_, err := d.deploy(ctx, spec)
			if err != nil && !test.err {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && test.err {
				t.Fatalf("expected an error")
			}

			require.Equal(t, test.expectedLogs, actionLogs.Bytes())

			output, err := os.ReadFile(outputFilePath)
			if test.expectedOutput == nil {
				require.ErrorIs(t, err, os.ErrNotExist)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedOutput, output)
			}

			test.appService.AssertExpectations(t)
		})
	}
}

type mockedRoundtripper struct {
	mock.Mock
}

func (m *mockedRoundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

type mockedAppsService struct {
	mock.Mock
	godo.AppsService
}

func (m *mockedAppsService) Get(ctx context.Context, appID string) (*godo.App, *godo.Response, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).(*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockedAppsService) Create(ctx context.Context, req *godo.AppCreateRequest) (*godo.App, *godo.Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockedAppsService) Update(ctx context.Context, name string, req *godo.AppUpdateRequest) (*godo.App, *godo.Response, error) {
	args := m.Called(ctx, name, req)
	return args.Get(0).(*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockedAppsService) List(ctx context.Context, opt *godo.ListOptions) ([]*godo.App, *godo.Response, error) {
	args := m.Called(ctx, opt)
	return args.Get(0).([]*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockedAppsService) GetDeployment(ctx context.Context, appID string, deploymentID string) (*godo.Deployment, *godo.Response, error) {
	args := m.Called(ctx, appID, deploymentID)
	return args.Get(0).(*godo.Deployment), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockedAppsService) ListDeployments(ctx context.Context, appID string, opt *godo.ListOptions) ([]*godo.Deployment, *godo.Response, error) {
	args := m.Called(ctx, appID, opt)
	return args.Get(0).([]*godo.Deployment), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockedAppsService) GetLogs(ctx context.Context, appID, deploymentID, component string, logType godo.AppLogType, follow bool, tailLines int) (*godo.AppLogs, *godo.Response, error) {
	args := m.Called(ctx, appID, deploymentID, component, logType, follow, tailLines)
	return args.Get(0).(*godo.AppLogs), args.Get(1).(*godo.Response), args.Error(2)
}
