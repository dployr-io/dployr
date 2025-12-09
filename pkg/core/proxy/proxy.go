// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"github.com/dployr-io/dployr/pkg/core/service"
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

// App describes a proxy by its domain, upstream service, root directory, proxy status and template type.
type App struct {
	Domain   string       `json:"domain"`
	Upstream string       `json:"upstream"`
	Root     string       `json:"root,omitempty"`
	Status   ProxyStatus  `json:"status"`
	Template TemplateType `json:"template"` // static, reverse_proxy, or php_fastcgi
}

// ProxyStatus describes the current status of the proxy service.
type ProxyStatus struct {
	Status service.SvcState `json:"status"`
}

// ProxyRoute describes a proxy by its domain and upstream service.
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
	Status() ProxyStatus

	// GetApps returns the current list of proxy apps
	GetApps() []App

	// Restart reloads Caddy with the current Caddyfile configuration.
	Restart() error

	// Add merges new apps with existing configuration, regenerates the Caddyfile,
	// and reloads Caddy. Duplicate domains are replaced with new configurations.
	Add(apps map[string]App) error

	// Remove deletes apps by domain, regenerates the Caddyfile, and reloads Caddy.
	Remove(domains []string) error
}
