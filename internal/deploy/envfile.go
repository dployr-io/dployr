// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dployr-io/dployr/pkg/store"
)

// WriteEnvFile writes <workDir>/.env from the blueprint's env vars and secrets,
// with PORT always written first. Duplicate keys are skipped — PORT wins over any
// env var named PORT, and env vars win over secrets with the same key.
func WriteEnvFile(workDir string, bp store.Blueprint, port int) error {
	var b strings.Builder
	written := map[string]bool{}

	write := func(k, v string) {
		if written[k] {
			return
		}
		written[k] = true
		fmt.Fprintf(&b, "%s=%s\n", k, v)
	}

	write("PORT", fmt.Sprintf("%d", port))
	for k, v := range bp.EnvVars {
		write(k, v)
	}
	for k, v := range bp.Secrets {
		write(k, v)
	}

	return os.WriteFile(filepath.Join(workDir, ".env"), []byte(b.String()), 0600)
}
