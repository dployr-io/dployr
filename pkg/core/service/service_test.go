// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package service

import "testing"

func TestSvcStateConstants(t *testing.T) {
	states := map[SvcState]string{
		SvcRunning: "running",
		SvcStopped: "stopped",
		SvcUnknown: "unknown",
	}

	for state, expected := range states {
		if string(state) != expected {
			t.Errorf("SvcState %q != %q", state, expected)
		}
	}
}
