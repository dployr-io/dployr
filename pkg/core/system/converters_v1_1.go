// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"time"

	"github.com/dployr-io/dployr/pkg/store"
)

// FromStoreDeployment converts a store.Deployment to the v1.1 format.
func FromStoreDeployment(d *store.Deployment) DeploymentV1_1 {
	if d == nil {
		return DeploymentV1_1{}
	}

	dep := DeploymentV1_1{
		ID:          d.ID,
		UserID:      d.UserId,
		Name:        d.Blueprint.Name,
		Description: d.Blueprint.Desc,
		Status:      string(d.Status),
		Source:      d.Blueprint.Source,
		Runtime: RuntimeInfo{
			Type: string(d.Blueprint.Runtime.Type),
		},
		Port:      d.Blueprint.Port,
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
		UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
	}

	if d.Blueprint.Runtime.Version != "" {
		dep.Runtime.Version = &d.Blueprint.Runtime.Version
	}

	if d.Blueprint.Remote.Url != "" {
		dep.Remote = &RemoteInfo{
			URL:        d.Blueprint.Remote.Url,
			Branch:     d.Blueprint.Remote.Branch,
			CommitHash: d.Blueprint.Remote.CommitHash,
		}
	}

	if d.Blueprint.WorkingDir != "" {
		dep.WorkingDir = &d.Blueprint.WorkingDir
	}
	if d.Blueprint.StaticDir != "" {
		dep.StaticDir = &d.Blueprint.StaticDir
	}
	if d.Blueprint.Image != "" {
		dep.Image = &d.Blueprint.Image
	}
	if d.Blueprint.RunCmd != "" {
		dep.RunCommand = &d.Blueprint.RunCmd
	}
	if d.Blueprint.BuildCmd != "" {
		dep.BuildCommand = &d.Blueprint.BuildCmd
	}

	if len(d.Blueprint.EnvVars) > 0 {
		dep.EnvVars = d.Blueprint.EnvVars
	}

	// Secrets - expose keys only with source indicator
	if len(d.Blueprint.Secrets) > 0 {
		dep.Secrets = make([]SecretRef, 0, len(d.Blueprint.Secrets))
		for key := range d.Blueprint.Secrets {
			dep.Secrets = append(dep.Secrets, SecretRef{
				Key:    key,
				Source: "local",
			})
		}
	}

	return dep
}

// FromStoreService converts a store.Service to the v1.1 format.
func FromStoreService(s *store.Service) ServiceV1_1 {
	if s == nil {
		return ServiceV1_1{}
	}

	svc := ServiceV1_1{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Runtime:     string(s.Runtime),
		Port:        s.Port,
		EnvVars:     make(map[string]string),
		Secrets:     []SecretRef{},
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}

	if s.RuntimeVersion != "" {
		svc.RuntimeVersion = &s.RuntimeVersion
	}
	if s.WorkingDir != "" {
		svc.WorkingDir = &s.WorkingDir
	}
	if s.StaticDir != "" {
		svc.StaticDir = &s.StaticDir
	}
	if s.Image != "" {
		svc.Image = &s.Image
	}
	if s.RunCmd != "" {
		svc.RunCommand = &s.RunCmd
	}
	if s.BuildCmd != "" {
		svc.BuildCommand = &s.BuildCmd
	}
	if s.Remote != "" {
		svc.RemoteURL = &s.Remote
	}
	if s.Branch != "" {
		svc.Branch = &s.Branch
	}
	if s.CommitHash != "" {
		svc.CommitHash = &s.CommitHash
	}

	if len(s.EnvVars) > 0 {
		svc.EnvVars = s.EnvVars
	}

	// Secrets - expose keys only with source indicator
	if len(s.Secrets) > 0 {
		svc.Secrets = make([]SecretRef, 0, len(s.Secrets))
		for key := range s.Secrets {
			svc.Secrets = append(svc.Secrets, SecretRef{
				Key:    key,
				Source: "local",
			})
		}
	}

	return svc
}

// FromStoreDeployments converts a slice of store.Deployment to v1.1 format.
func FromStoreDeployments(deployments []*store.Deployment) []DeploymentV1_1 {
	if len(deployments) == 0 {
		return nil
	}

	result := make([]DeploymentV1_1, 0, len(deployments))
	for _, d := range deployments {
		result = append(result, FromStoreDeployment(d))
	}
	return result
}

// FromStoreServices converts a slice of store.Service to v1.1 format.
func FromStoreServices(services []*store.Service) []ServiceV1_1 {
	if len(services) == 0 {
		return nil
	}

	result := make([]ServiceV1_1, 0, len(services))
	for _, s := range services {
		result = append(result, FromStoreService(s))
	}
	return result
}
