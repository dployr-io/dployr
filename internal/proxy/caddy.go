// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/service"
	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
)

//go:embed templates/*.tpl
var templateFS embed.FS

type CaddyHandler struct {
	Apps   map[string]proxy.App
	logger *shared.Logger
}

func Init(a map[string]proxy.App, logger *shared.Logger) *CaddyHandler {
	return &CaddyHandler{
		Apps:   a,
		logger: logger,
	}
}

const stateFile = ".dployr/caddy/apps.json"

type TemplateData struct {
	Apps    map[string]proxy.App
	LogDir  string
	HomeDir string
	Content string
}

type AppTemplateData struct {
	Domain  string
	App     proxy.App
	LogDir  string
	LogFile string
	HomeDir string
}

func (c *CaddyHandler) Setup(apps map[string]proxy.App) error {
	c.logger.Debug("setup called", "app_count", len(apps))
	for domain, app := range apps {
		c.logger.Debug("app loaded", "domain", domain, "upstream", app.Upstream, "root", app.Root, "template", app.Template)
	}

	// Parse embedded templates
	tmpl, err := template.ParseFS(templateFS, "templates/*.tpl")
	if err != nil {
		return fmt.Errorf("unable to parse embedded templates: %w", err)
	}

	dataDir := utils.GetDataDir()
	cfgDir := filepath.Join(dataDir, ".dployr", "caddy")
	err = os.MkdirAll(cfgDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create config directory: %w", err)
	}

	// create log directory
	logDir := filepath.Join(dataDir, ".dployr", "logs", "caddy")
	c.logger.Debug("creating log directory", "path", logDir)
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create log directory: %w", err)
	}
	c.logger.Debug("log directory created")

	out, err := os.Create(filepath.Join(cfgDir, "Caddyfile"))
	if err != nil {
		return fmt.Errorf("unable to create Caddyfile: %w", err)
	}
	defer out.Close()

	// generate content for each app
	var contentBuilder strings.Builder
	for domain, app := range apps {
		logFile := filepath.Join(logDir, domain+".log")
		appData := AppTemplateData{
			Domain:  domain,
			App:     app,
			LogDir:  logDir,
			LogFile: logFile,
			HomeDir: dataDir,
		}

		templateName := fmt.Sprintf("%s.tpl", app.Template)
		c.logger.Debug("processing app", "domain", domain, "template", templateName)

		if err := tmpl.ExecuteTemplate(&contentBuilder, templateName, appData); err != nil {
			c.logger.Error("template execution failed", "template", templateName, "domain", domain, "error", err)
			return fmt.Errorf("unable to execute template %s for app %s: %w", templateName, domain, err)
		}
		contentBuilder.WriteString("\n")
	}

	tmplData := TemplateData{
		Apps:    apps,
		LogDir:  logDir,
		HomeDir: dataDir,
		Content: contentBuilder.String(),
	}

	c.logger.Debug("executing caddyfile template", "app_count", len(tmplData.Apps), "content_length", len(tmplData.Content))

	if err := tmpl.ExecuteTemplate(out, "caddyfile.tpl", tmplData); err != nil {
		return fmt.Errorf("unable to execute template: %w", err)
	}

	c.logger.Debug("caddyfile written", "path", filepath.Join(cfgDir, "Caddyfile"))

	// Save state
	return saveState(apps)
}

func (c *CaddyHandler) Stop() error {
	c.logger.Info("stopping caddy service")
	stop := exec.Command("sudo", "systemctl", "stop", "caddy")
	if output, err := stop.CombinedOutput(); err != nil {
		c.logger.Warn("failed to stop caddy service", "output", string(output), "error", err)
		return fmt.Errorf("failed to stop caddy service: %w", err)
	}
	c.logger.Info("caddy service stopped")
	return nil
}

func (c *CaddyHandler) Status() proxy.ProxyStatus {
	resp, err := http.Get("http://localhost:2019/config/")
	if err != nil {
		return proxy.ProxyStatus{
			Status: service.SvcStopped,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var config map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
			c.logger.Warn("failed to decode caddy config", "error", err)
			return proxy.ProxyStatus{
				Status: service.SvcRunning,
			}
		}

		status := proxy.ProxyStatus{
			Status: service.SvcRunning,
		}

		if apps, ok := config["apps"]; ok {
			status.Apps = apps.(map[string]any)
		}

		return status
	}
	return proxy.ProxyStatus{
		Status: service.SvcUnknown,
	}
}

func (c *CaddyHandler) GetApps() []proxy.App {
	state := LoadState()
	apps := make([]proxy.App, 0, len(state))
	for _, app := range state {
		apps = append(apps, app)
	}
	return apps
}

func (c *CaddyHandler) Restart() error {
	dataDir := utils.GetDataDir()
	cfgPath := filepath.Join(dataDir, ".dployr", "caddy", "Caddyfile")

	c.logger.Debug("validating caddy config", "path", cfgPath)
	validate := exec.Command("sudo", "caddy", "validate", "--config", cfgPath, "--adapter", "caddyfile")
	res, err := validate.CombinedOutput()
	if err != nil {
		return fmt.Errorf("caddy config validation failed: %w -> %s", err, string(res))
	}
	c.logger.Debug("caddy config validation successful")

	c.logger.Info("restarting caddy service")
	restart := exec.Command("sudo", "systemctl", "restart", "caddy")
	if output, err := restart.CombinedOutput(); err != nil {
		c.logger.Error("failed to restart caddy service", "output", string(output), "error", err)
		return fmt.Errorf("failed to restart caddy service: %w", err)
	}

	// verify it's actually running by checking admin API
	maxRetries := 10
	for i := range maxRetries {
		time.Sleep(200 * time.Millisecond)
		status := c.Status()
		if status.Status == service.SvcRunning {
			c.logger.Info("caddy service is running and responding")
			return nil
		}
		c.logger.Debug("waiting for caddy to respond", "attempt", i+1, "max", maxRetries)
	}

	return fmt.Errorf("caddy service started but is not responding on admin port after %d retries", maxRetries)
}

func (c *CaddyHandler) Add(apps map[string]proxy.App) error {
	existing := LoadState()
	c.logger.Debug("adding apps", "previous_count", len(existing))

	for domain, app := range apps {
		existing[domain] = app
		c.logger.Debug("merged app", "domain", domain)
	}

	c.logger.Debug("apps merged", "new_count", len(existing))
	if err := c.Setup(existing); err != nil {
		return err
	}

	return c.Restart()
}

func (c *CaddyHandler) Remove(domains []string) error {
	existing := LoadState()

	for _, domain := range domains {
		delete(existing, domain)
	}

	if err := c.Setup(existing); err != nil {
		return err
	}

	return c.Restart()
}

// LoadState reads app config from apps.json.
// Returns an empty map if the file doesn't exist.
func LoadState() map[string]proxy.App {
	dataDir := utils.GetDataDir()
	statePath := filepath.Join(dataDir, stateFile)

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return make(map[string]proxy.App)
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil
	}

	var apps map[string]proxy.App
	if err := json.Unmarshal(data, &apps); err != nil {
		return nil
	}

	return apps
}

// saveState persists app config to apps.json
func saveState(apps map[string]proxy.App) error {
	dataDir := utils.GetDataDir()

	data, err := json.MarshalIndent(apps, "", "  ")
	if err != nil {
		return err
	}

	statePath := filepath.Join(dataDir, stateFile)
	return os.WriteFile(statePath, data, 0644)
}
