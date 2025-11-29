// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
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
)

//go:embed templates/*.tpl
var templateFS embed.FS

type CaddyHandler struct {
	Apps    map[string]proxy.App
	process *os.Process
}

func Init(a map[string]proxy.App) *CaddyHandler {
	return &CaddyHandler{
		Apps: a,
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
	// [DEBUG]
	log.Printf("setup called with %d apps", len(apps))
	for domain, app := range apps {
		log.Printf("app: %s -> %+v", domain, app)
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
	log.Printf("Creating log directory: %s", logDir)
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create log directory: %w", err)
	}
	log.Printf("log directory created successfully")

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
		log.Printf("processing app %s with template %s", domain, templateName)

		if err := tmpl.ExecuteTemplate(&contentBuilder, templateName, appData); err != nil {
			log.Printf("template execution failed: template=%s, domain=%s, error=%v", templateName, domain, err)
			log.Printf("available templates: %v", tmpl.DefinedTemplates())
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

	// [DEBUG]
	log.Printf("template data: Apps=%d, LogDir=%s, Content length=%d", len(tmplData.Apps), tmplData.LogDir, len(tmplData.Content))
	log.Printf("generated content:\n%s", tmplData.Content)

	if err := tmpl.ExecuteTemplate(out, "caddyfile.tpl", tmplData); err != nil {
		return fmt.Errorf("unable to execute template: %w", err)
	}

	// [DEBUG]
	log.Printf("caddyfile written to: %s", filepath.Join(cfgDir, "Caddyfile"))

	// Save state
	return saveState(apps)
}

func (c *CaddyHandler) Stop() error {
	stop := exec.Command("caddy", "stop")
	if output, err := stop.CombinedOutput(); err != nil {
		log.Printf("attempt to stop caddy failed: %s", string(output))

		if c.process != nil {
			log.Printf("terminating caddy process PID: %d", c.process.Pid)
			if err := c.process.Kill(); err != nil {
				return fmt.Errorf("failed to kill caddy process: %w", err)
			}
			c.process = nil
		}
	}

	return nil
}

func (c *CaddyHandler) Status() proxy.ProxyStatusResponse {
	resp, err := http.Get("http://localhost:2019/config/")
	if err != nil {
		return proxy.ProxyStatusResponse{
			Status: service.SvcStopped,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return proxy.ProxyStatusResponse{
			Status: service.SvcRunning,
		}
	}
	return proxy.ProxyStatusResponse{
		Status: service.SvcUnknown,
	}
}

func (c *CaddyHandler) Restart() error {
	dataDir := utils.GetDataDir()
	cfgPath := filepath.Join(dataDir, ".dployr", "caddy", "Caddyfile")

	// stop first to avoid port conflicts
	log.Printf("stopping any existing caddy instance")
	if err := c.Stop(); err != nil {
		log.Printf("error stopping caddy: %v", err)
	}

	// wait a moment for port to be released
	time.Sleep(100 * time.Millisecond)

	log.Printf("validating caddy config: %s", cfgPath)
	validate := exec.Command("caddy", "validate", "--config", cfgPath)
	res, err := validate.CombinedOutput()
	if err != nil {
		return fmt.Errorf("caddy config validation failed: %w -> %s", err, string(res))
	}
	log.Printf("caddy config validation successful")

	cmd := exec.Command("caddy", "run", "--config", cfgPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("caddy run failed to start: %w", err)
	}
	c.process = cmd.Process
	log.Printf("caddy process started with PID: %d", c.process.Pid)

	// verify it's actually running by checking admin API
	maxRetries := 3
	for range maxRetries {
		time.Sleep(100 * time.Millisecond)
		status := c.Status()
		if status.Status == service.SvcRunning {
			log.Printf("caddy is running and responding")
			return nil
		}
	}

	return fmt.Errorf("caddy started but is not responding on admin port after %d retries", maxRetries)
}

func (c *CaddyHandler) Add(apps map[string]proxy.App) error {
	existing := LoadState()
	// [DEBUG]
	log.Printf("previous count: %d", len(existing))

	for domain, app := range apps {
		existing[domain] = app
		log.Printf("merged app %s into existing", domain)
	}

	// [DEBUG]
	log.Printf("new count: %d", len(existing))
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
	log.Printf("checking state file at %s", statePath)

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		log.Printf("no state file was found, returning empty map")
		return make(map[string]proxy.App)
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		log.Printf("error reading state file: %v", err)
		return nil
	}

	var apps map[string]proxy.App
	if err := json.Unmarshal(data, &apps); err != nil {
		log.Printf("error unmarshaling: %v", err)
		return nil
	}

	// [DEBUG]
	log.Printf("loaded %d apps from state file", len(apps))
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
