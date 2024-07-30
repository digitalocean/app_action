package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindAppByName(t *testing.T) {
	app1 := &godo.App{Spec: &godo.AppSpec{Name: "app1"}}
	app2 := &godo.App{Spec: &godo.AppSpec{Name: "app2"}}

	as := &mockedAppsService{}
	as.On("List", mock.Anything, &godo.ListOptions{Page: 0}).Return([]*godo.App{app1}, &godo.Response{Links: &godo.Links{Pages: &godo.Pages{Next: "2"}}}, nil).Times(3)
	as.On("List", mock.Anything, &godo.ListOptions{Page: 2}).Return([]*godo.App{app2}, &godo.Response{}, nil).Times(2)

	app, err := FindAppByName(context.Background(), as, "app1")
	require.NoError(t, err)
	require.Equal(t, app1, app)

	app, err = FindAppByName(context.Background(), as, "app2")
	require.NoError(t, err)
	require.Equal(t, app2, app)

	app, err = FindAppByName(context.Background(), as, "app3")
	require.NoError(t, err)
	require.Nil(t, app)

	as.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, &godo.Response{}, errors.New("an error")).Once()
	_, err = FindAppByName(context.Background(), as, "app4")
	require.Error(t, err)

	as.AssertExpectations(t)
}

type mockedAppsService struct {
	godo.AppsService
	mock.Mock
}

func (m *mockedAppsService) List(ctx context.Context, opt *godo.ListOptions) ([]*godo.App, *godo.Response, error) {
	args := m.Called(ctx, opt)
	return args.Get(0).([]*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}
