package shared

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Exec(ctx context.Context, cmd string, workDir string) error {
	var shell, shellFlag, setup string

	switch runtime.GOOS {
	case "windows":
		shell = "pwsh"
		shellFlag = "-Command"
		setup = `Invoke-Expression "$(vfox activate pwsh)"`
	default:
		shell = "bash"
		shellFlag = "-lc"
		setup = `eval "$(vfox activate bash)"`
	}

	full := fmt.Sprintf("%s && cd %s && %s", setup, workDir, cmd)

	c := exec.CommandContext(ctx, shell, shellFlag, full)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = os.Environ() 

	return c.Run()
}
