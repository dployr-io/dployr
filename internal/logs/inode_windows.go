//go:build windows

// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package logs

// getInode is not supported on Windows.
func getInode(sys interface{}) uint64 {
	return 0
}
