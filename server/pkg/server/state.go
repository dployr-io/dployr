package server

import (
	"fmt"
	"time"

	"dployr.io/pkg/models"
)

type SetupState struct {
	Version string `json:"version"`
	ProjectId string `json:"project_id"`
	ServerId string `json:"server_id"`
	Status string `json:"status"` // success | pending | failed
	Phases Phases `json:"phases"`
	CreatedAt time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
}

type Phases struct {
	SystemDeps SystemDeps `json:"system_deps"`
	Services Services `json:"services"`
	FireWall FireWall `json:"firewall"`
	SslSetup SslSetup `json:"ssl_setup"`
}

type SystemDeps struct {
	Status string `json:"status"` // success | pending | failed
	NodejsVersion string `json:"nodejs_version"` // v20
}

type Services struct {
	Status string `json:"status"` // success | pending | failed
	NginxInstalled bool `json:"nginx_installed"`
	PM2Installed bool `json:"pm2_installed"`
}

type FireWall struct {
	Status string `json:"status"` // success | pending | failed
	PortsOpen []int `json:"ports_open"`
}

type SslSetup struct {
	Status string `json:"status"` // success | pending | failed
	Domain string `json:"domain"`
	CertPathLive string `json:"cert_path_live"`
}

func NewSetupState(p *models.Project) *SetupState {
	return &SetupState{
		ProjectId: fmt.Sprintf("proj_%d_%s", p.Name,time.Now().Unix()),
		ServerId: fmt.Sprintf("srv_%s_%d_%s", p.Name, time.Now().Unix()),
		Status: models.Pending,
		Phases: Phases{},
	}
}


