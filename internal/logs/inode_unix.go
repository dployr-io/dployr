//go:build !windows

// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package logs

import "syscall"

// getInode extracts inode number from os.FileInfo.Sys() (Unix-like systems only).
func getInode(sys interface{}) uint64 {
	if stat, ok := sys.(*syscall.Stat_t); ok {
		return stat.Ino
	}
	return 0
}
