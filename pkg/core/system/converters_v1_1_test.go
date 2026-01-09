// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"testing"
	"time"

	"github.com/dployr-io/dployr/pkg/store"
)

func TestFromStoreDeployment(t *testing.T) {
	now := time.Now()
	userID := "user123"
	version := "1.20"
	workingDir := "/app"
	staticDir := "/static"
	image := "nginx:latest"
	runCmd := "go run main.go"
	buildCmd := "go build"

	deployment := &store.Deployment{
		ID:     "dep-001",
		UserId: &userID,
		Blueprint: store.Blueprint{
			Name:       "test-app",
			Desc:       "Test deployment",
			Source:     "remote",
			Runtime:    store.RuntimeObj{Type: store.RuntimeGo, Version: version},
			Remote:     store.RemoteObj{Url: "https://github.com/test/repo", Branch: "main", CommitHash: "abc123"},
			Port:       8080,
			WorkingDir: workingDir,
			StaticDir:  staticDir,
			Image:      image,
			RunCmd:     runCmd,
			BuildCmd:   buildCmd,
			EnvVars:    map[string]string{"ENV": "prod"},
			Secrets:    map[string]string{"API_KEY": "secret123"},
		},
		Status:    store.StatusCompleted,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := FromStoreDeployment(deployment)

	if result.ID != "dep-001" {
		t.Errorf("Expected ID 'dep-001', got '%s'", result.ID)
	}
	if result.UserID == nil || *result.UserID != userID {
		t.Errorf("Expected UserID '%s', got %v", userID, result.UserID)
	}
	if result.Name != "test-app" {
		t.Errorf("Expected Name 'test-app', got '%s'", result.Name)
	}
	if result.Status != "completed" {
		t.Errorf("Expected Status 'completed', got '%s'", result.Status)
	}
	if result.Runtime.Type != "golang" {
		t.Errorf("Expected Runtime.Type 'golang', got '%s'", result.Runtime.Type)
	}
	if result.Runtime.Version == nil || *result.Runtime.Version != version {
		t.Errorf("Expected Runtime.Version '%s', got %v", version, result.Runtime.Version)
	}
	if result.Remote == nil {
		t.Fatal("Expected Remote to be non-nil")
	}
	if result.Remote.URL != "https://github.com/test/repo" {
		t.Errorf("Expected Remote.URL 'https://github.com/test/repo', got '%s'", result.Remote.URL)
	}
	if result.Port != 8080 {
		t.Errorf("Expected Port 8080, got %d", result.Port)
	}
	if result.WorkingDir == nil || *result.WorkingDir != workingDir {
		t.Errorf("Expected WorkingDir '%s', got %v", workingDir, result.WorkingDir)
	}
	if len(result.EnvVars) != 1 || result.EnvVars["ENV"] != "prod" {
		t.Errorf("Expected EnvVars map with ENV=prod, got %v", result.EnvVars)
	}
	if len(result.Secrets) != 1 || result.Secrets[0].Key != "API_KEY" {
		t.Errorf("Expected Secrets with API_KEY, got %v", result.Secrets)
	}
}

func TestFromStoreDeployment_Nil(t *testing.T) {
	result := FromStoreDeployment(nil)
	if result.ID != "" {
		t.Errorf("Expected empty DeploymentV1_1 for nil input, got %+v", result)
	}
}

func TestFromStoreService(t *testing.T) {
	now := time.Now()
	runtimeVersion := "3.11"
	workingDir := "/app"
	remote := "https://github.com/test/service"
	branch := "develop"
	commitHash := "def456"

	service := &store.Service{
		ID:             "svc-001",
		Name:           "api-service",
		Description:    "API Service",
		Source:         "remote",
		Runtime:        store.RuntimePython,
		RuntimeVersion: runtimeVersion,
		Port:           3000,
		WorkingDir:     workingDir,
		Remote:         remote,
		Branch:         branch,
		CommitHash:     commitHash,
		EnvVars:        map[string]string{"DEBUG": "true"},
		Secrets:        map[string]string{"DB_PASS": "secret"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := FromStoreService(service)

	if result.ID != "svc-001" {
		t.Errorf("Expected ID 'svc-001', got '%s'", result.ID)
	}
	if result.Name != "api-service" {
		t.Errorf("Expected Name 'api-service', got '%s'", result.Name)
	}
	if result.Runtime != "python" {
		t.Errorf("Expected Runtime 'python', got '%s'", result.Runtime)
	}
	if result.RuntimeVersion == nil || *result.RuntimeVersion != runtimeVersion {
		t.Errorf("Expected RuntimeVersion '%s', got %v", runtimeVersion, result.RuntimeVersion)
	}
	if result.Port != 3000 {
		t.Errorf("Expected Port 3000, got %d", result.Port)
	}
	if result.WorkingDir == nil || *result.WorkingDir != workingDir {
		t.Errorf("Expected WorkingDir '%s', got %v", workingDir, result.WorkingDir)
	}
	if result.RemoteURL == nil || *result.RemoteURL != remote {
		t.Errorf("Expected RemoteURL '%s', got %v", remote, result.RemoteURL)
	}
	if result.Branch == nil || *result.Branch != branch {
		t.Errorf("Expected Branch '%s', got %v", branch, result.Branch)
	}
	if len(result.EnvVars) != 1 || result.EnvVars["DEBUG"] != "true" {
		t.Errorf("Expected EnvVars map with DEBUG=true, got %v", result.EnvVars)
	}
	if len(result.Secrets) != 1 || result.Secrets[0].Key != "DB_PASS" {
		t.Errorf("Expected Secrets with DB_PASS, got %v", result.Secrets)
	}
}

func TestFromStoreService_Nil(t *testing.T) {
	result := FromStoreService(nil)
	if result.ID != "" {
		t.Errorf("Expected empty ServiceV1_1 for nil input, got %+v", result)
	}
}

func TestFromStoreDeployments(t *testing.T) {
	now := time.Now()
	userID := "user123"

	deployments := []*store.Deployment{
		{
			ID:     "dep-001",
			UserId: &userID,
			Blueprint: store.Blueprint{
				Name:    "app1",
				Runtime: store.RuntimeObj{Type: store.RuntimeGo},
				Port:    8080,
			},
			Status:    store.StatusCompleted,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:     "dep-002",
			UserId: &userID,
			Blueprint: store.Blueprint{
				Name:    "app2",
				Runtime: store.RuntimeObj{Type: store.RuntimeNodeJS},
				Port:    3000,
			},
			Status:    store.StatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	result := FromStoreDeployments(deployments)

	if len(result) != 2 {
		t.Fatalf("Expected 2 deployments, got %d", len(result))
	}
	if result[0].ID != "dep-001" {
		t.Errorf("Expected first deployment ID 'dep-001', got '%s'", result[0].ID)
	}
	if result[1].ID != "dep-002" {
		t.Errorf("Expected second deployment ID 'dep-002', got '%s'", result[1].ID)
	}
}

func TestFromStoreDeployments_Empty(t *testing.T) {
	result := FromStoreDeployments(nil)
	if result != nil {
		t.Errorf("Expected nil for empty input, got %v", result)
	}

	result = FromStoreDeployments([]*store.Deployment{})
	if result != nil {
		t.Errorf("Expected nil for empty slice, got %v", result)
	}
}

func TestFromStoreServices(t *testing.T) {
	now := time.Now()

	services := []*store.Service{
		{
			ID:        "svc-001",
			Name:      "service1",
			Runtime:   store.RuntimePython,
			Port:      5000,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "svc-002",
			Name:      "service2",
			Runtime:   store.RuntimeRuby,
			Port:      4000,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	result := FromStoreServices(services)

	if len(result) != 2 {
		t.Fatalf("Expected 2 services, got %d", len(result))
	}
	if result[0].ID != "svc-001" {
		t.Errorf("Expected first service ID 'svc-001', got '%s'", result[0].ID)
	}
	if result[1].ID != "svc-002" {
		t.Errorf("Expected second service ID 'svc-002', got '%s'", result[1].ID)
	}
}

func TestFromStoreServices_Empty(t *testing.T) {
	result := FromStoreServices(nil)
	if result != nil {
		t.Errorf("Expected nil for empty input, got %v", result)
	}

	result = FromStoreServices([]*store.Service{})
	if result != nil {
		t.Errorf("Expected nil for empty slice, got %v", result)
	}
}
