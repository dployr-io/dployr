// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package system

import (
	"io/fs"
)

func getOwnershipPlatform(info fs.FileInfo) (uid, gid int, owner, group string) {
	// Windows doesn't have Unix-style UID/GID
	// Return placeholder values
	return 0, 0, "SYSTEM", "SYSTEM"
}
