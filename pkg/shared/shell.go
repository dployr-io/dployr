// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Exec(ctx context.Context, cmd string, workDir string) error {
	return ExecWithOptions(ctx, cmd, workDir, false)
}

func ExecWithOptions(ctx context.Context, cmd string, workDir string, useVfox bool) error {
	var shell, shellFlag, full string

	switch runtime.GOOS {
	case "windows":
		shell = "pwsh"
		shellFlag = "-Command"
		if useVfox {
			full = fmt.Sprintf(`Invoke-Expression "$(vfox activate pwsh)"; cd %s; %s`, workDir, cmd)
		} else {
			full = fmt.Sprintf(`cd %s; %s`, workDir, cmd)
		}
	default:
		shell = "bash"
		shellFlag = "-lc"
		if useVfox {
			full = fmt.Sprintf(`eval "$(vfox activate bash)" && cd %s && %s`, workDir, cmd)
		} else {
			full = fmt.Sprintf(`cd %s && %s`, workDir, cmd)
		}
	}

	c := exec.CommandContext(ctx, shell, shellFlag, full)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = os.Environ()

	return c.Run()
}
