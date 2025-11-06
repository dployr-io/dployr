package shared

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"context"
)

func Exec(ctx context.Context, cmd string, workDir string) error {
	var shell, shellFlag, setup string

	switch runtime.GOOS {
	case "windows":
		shell = "pwsh"
		shellFlag = "-Command"
		setup = `if (-not (Test-Path -Path $PROFILE)) { New-Item -Type File -Path $PROFILE -Force }; Add-Content -Path $PROFILE -Value 'Invoke-Expression "$(vfox activate pwsh)"'`
	default:
		shell = "bash"
		shellFlag = "-lc"
		setup = `eval "$(vfox activate bash)"` // works on both macOS and Linux
	}

	full := fmt.Sprintf("%s && cd %s && %s", setup, workDir, cmd)

	c := exec.CommandContext(ctx, shell, shellFlag, full)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = append(os.Environ(),
		fmt.Sprintf("VFOX_HOME=%s/.version-fox", os.Getenv("HOME")),
	)

	return c.Run()
}

