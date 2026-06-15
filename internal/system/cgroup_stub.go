// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package system

import "github.com/dployr-io/dployr/pkg/core/system"

func EnsureClusterSlice(clusterID string, memoryMB int, cpuMillicores int) error {
	return nil
}

func ReadClusterResources() map[string]*system.ClusterResourcesInfo {
	return nil
}
