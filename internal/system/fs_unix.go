// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package system

import (
	"io/fs"
	"syscall"
)

func getOwnershipPlatform(info fs.FileInfo) (uid, gid int, owner, group string) {
	uid = -1
	gid = -1
	owner = "unknown"
	group = "unknown"

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		uid = int(stat.Uid)
		gid = int(stat.Gid)
		owner = lookupUser(uid)
		group = lookupGroup(gid)
	}

	return
}
