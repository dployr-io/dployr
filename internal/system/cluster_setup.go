// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

// ClusterSetup implements cluster.Setupper using EnsureClusterSlice.
type ClusterSetup struct{}

func NewClusterSetup() *ClusterSetup { return &ClusterSetup{} }

func (c *ClusterSetup) Setup(clusterID string, memoryMB int, cpuMillicores int) error {
	return EnsureClusterSlice(clusterID, memoryMB, cpuMillicores)
}
