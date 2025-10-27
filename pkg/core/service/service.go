package service

type SvcState string

const (
	SvcRunning SvcState = "running"
	SvcStopped SvcState = "stopped"
	SvcUnknown SvcState = "unknown"
)

// type SvcMgr interface {
// 	Status(name string) (string, error)
// 	Install(name, desc, runCmd, workDir string, envVars map[string]string) error
// 	Start(name string) error
// 	Stop(name string) error
// 	Remove(name string) error
// }

// func GetSvcMgr() (SvcMgr, error) {
// 	switch runtime.GOOS {
// 	case "windows":
// 		return &NSSMManager{}, nil
// 	case "linux":
// 		return nil, fmt.Errorf("systemd manager not yet implemented")
// 	case "darwin":
// 		return nil, fmt.Errorf("launchd manager not yet implemented")
// 	default:
// 		return nil, fmt.Errorf("unsupported platform")
// 	}
// }
