// +build integration

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ssm_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/store"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/store/ssm"
	"github.com/stretchr/testify/require"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Test_SSM_Project_Integration(t *testing.T) {
	s, _ := ssm.NewStore()
	projectToCreate := archer.Project{Name: randStringBytes(10), Version: "1.0"}
	t.Run("Create, Get and List Projects", func(t *testing.T) {
		// Create our first project
		err := s.CreateProject(&projectToCreate)
		require.NoError(t, err)

		// Can't overwrite an existing project
		err = s.CreateProject(&projectToCreate)
		require.EqualError(t, &store.ErrProjectAlreadyExists{ProjectName: projectToCreate.Name}, err.Error())

		// Fetch the project back from SSM
		project, err := s.GetProject(projectToCreate.Name)
		require.NoError(t, err)
		require.Equal(t, projectToCreate, *project)

		// List returns a non-empty list of projects
		projects, err := s.ListProjects()
		require.NoError(t, err)
		require.NotEmpty(t, projects)
	})
}

func Test_SSM_Environment_Integration(t *testing.T) {
	s, _ := ssm.NewStore()
	projectToCreate := archer.Project{Name: randStringBytes(10), Version: "1.0"}
	testEnvironment := archer.Environment{Name: "test", Project: projectToCreate.Name, Region: "us-west-2", AccountID: " 1234"}
	prodEnvironment := archer.Environment{Name: "prod", Project: projectToCreate.Name, Region: "us-west-2", AccountID: " 1234"}

	t.Run("Create, Get and List Environments", func(t *testing.T) {
		// Create our first project
		err := s.CreateProject(&projectToCreate)
		require.NoError(t, err)

		// Make sure there are no envs with our new project
		envs, err := s.ListEnvironments(projectToCreate.Name)
		require.NoError(t, err)
		require.Empty(t, envs)

		// Add our environments
		err = s.CreateEnvironment(&testEnvironment)
		require.NoError(t, err)

		err = s.CreateEnvironment(&prodEnvironment)
		require.NoError(t, err)

		// Make sure we can't add a duplicate environment
		err = s.CreateEnvironment(&prodEnvironment)
		require.EqualError(t, &store.ErrEnvironmentAlreadyExists{ProjectName: projectToCreate.Name, EnvironmentName: prodEnvironment.Name}, err.Error())

		// Wait for consistency to kick in (ssm path commands are eventually consistent)
		time.Sleep(5 * time.Second)

		// Make sure all the environments are under our project
		envs, err = s.ListEnvironments(projectToCreate.Name)
		require.NoError(t, err)
		var environments []archer.Environment
		for _, e := range envs {
			environments = append(environments, *e)
		}
		require.ElementsMatch(t, environments, []archer.Environment{testEnvironment, prodEnvironment})

		// Fetch our saved environments, one by one
		env, err := s.GetEnvironment(projectToCreate.Name, testEnvironment.Name)
		require.NoError(t, err)
		require.Equal(t, testEnvironment, *env)

		env, err = s.GetEnvironment(projectToCreate.Name, prodEnvironment.Name)
		require.NoError(t, err)
		require.Equal(t, prodEnvironment, *env)
	})
}

func Test_SSM_Application_Integration(t *testing.T) {
	s, _ := ssm.NewStore()
	projectToCreate := archer.Project{Name: randStringBytes(10), Version: "1.0"}
	apiApplication := archer.Application{Name: "api", Project: projectToCreate.Name, Type: "LBFargateService"}
	feApplication := archer.Application{Name: "front-end", Project: projectToCreate.Name, Type: "LBFargateService"}

	t.Run("Create, Get and List Applications", func(t *testing.T) {
		// Create our first project
		err := s.CreateProject(&projectToCreate)
		require.NoError(t, err)

		// Make sure there are no apps with our new project
		apps, err := s.ListApplications(projectToCreate.Name)
		require.NoError(t, err)
		require.Empty(t, apps)

		// Add our applications
		err = s.CreateApplication(&apiApplication)
		require.NoError(t, err)

		err = s.CreateApplication(&feApplication)
		require.NoError(t, err)

		// Make sure we can't add a duplicate apps
		err = s.CreateApplication(&apiApplication)
		require.EqualError(t, &store.ErrApplicationAlreadyExists{ProjectName: projectToCreate.Name, ApplicationName: apiApplication.Name}, err.Error())

		// Wait for consistency to kick in (ssm path commands are eventually consistent)
		time.Sleep(5 * time.Second)

		// Make sure all the apps are under our project
		apps, err = s.ListApplications(projectToCreate.Name)
		require.NoError(t, err)
		var applications []archer.Application
		for _, a := range apps {
			applications = append(applications, *a)
		}
		require.ElementsMatch(t, applications, []archer.Application{apiApplication, feApplication})

		// Fetch our saved apps, one by one
		app, err := s.GetApplication(projectToCreate.Name, apiApplication.Name)
		require.NoError(t, err)
		require.Equal(t, apiApplication, *app)

		app, err = s.GetApplication(projectToCreate.Name, feApplication.Name)
		require.NoError(t, err)
		require.Equal(t, feApplication, *app)
	})
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
