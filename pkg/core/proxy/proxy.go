package proxy

import (
	"dployr/pkg/core/service"
)

type TemplateType string

const (
	TemplateStatic       TemplateType = "static"
	TemplateReverseProxy TemplateType = "reverse_proxy"
	TemplatePHPFastCGI   TemplateType = "php_fastcgi"
)

type Proxier struct {
	apps map[string]App
	api  HandleProxy
}

func NewProxier(a map[string]App, p HandleProxy) *Proxier {
	return &Proxier{
		apps: a,
		api:  p,
	}
}

type App struct {
	Domain   string       `json:"domain"`
	Upstream string       `json:"upstream"`
	Root     string       `json:"root,omitempty"`
	Template TemplateType `json:"template"` // static, reverse_proxy, or php_fastcgi
}

type ProxyStatusResponse struct {
	Status service.SvcState `json:"status"`
}

type ProxyStatus struct {
	Status string `json:"status"`
}

type ProxyRoute struct {
	Domain   string `json:"domain"`
	Upstream string `json:"upstream"`
}

type HandleProxy interface {
	// Setup generates a Caddyfile from templates and persists app configuration state.
	// It creates the config directory if needed and saves all apps to apps.json for
	// later modification operations.
	Setup(apps map[string]App) error

	// Status returns the current status of the proxy service
	Status() ProxyStatusResponse

	// Restart reloads Caddy with the current Caddyfile configuration.
	Restart() error

	// Add merges new apps with existing configuration, regenerates the Caddyfile,
	// and reloads Caddy. Duplicate domains are replaced with new configurations.
	Add(apps map[string]App) error

	// Remove deletes apps by domain, regenerates the Caddyfile, and reloads Caddy.
	Remove(domains []string) error
}
