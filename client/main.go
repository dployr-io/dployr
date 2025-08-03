package main

import (
	"context"
	"embed"

	"dployr/core/data"
	"dployr/core/domain"
	"dployr/core/terminal"

	"dployr.io/pkg/models"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx             context.Context
	dataService     *data.DataService
	domainService   *domain.DomainService
	terminalService *terminal.TerminalService
}

func NewApp() *App {
	return &App{}
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "dployr desktop",
		MinWidth:  1024,
		MinHeight: 768,

		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.dataService = data.NewDataService()
	a.domainService = domain.NewDomainService(getBaseUrl())
	a.terminalService = terminal.NewTerminalService(ctx)
}

func getBaseUrl() string {
	return ""
}

// Wails binding methods - delegate to services
func (a *App) SignIn(host, email, name string) (*models.MessageResponse, error) {
	return a.dataService.SignIn(host, email, name)
}

func (a *App) VerifyMagicCode(host, email, code string) (*models.AuthTokenPair, error) {
	return a.dataService.VerifyMagicCode(host, email, code)
}

func (a *App) GetDeployments(host, token, refresh, projectId string) ([]models.Deployment, error) {
	return a.dataService.GetDeployments(host, token, refresh, projectId)
}

func (a *App) GetLogs(host, token, refresh, projectId string) ([]models.LogEntry, error) {
	return a.dataService.GetLogs(host, token, refresh, projectId)
}

func (a *App) GetProjects(host, token string) ([]models.Project, error) {
	return a.dataService.GetProjects(host, token)
}

func (a *App) CreateProject(host, token string, payload map[string]string) (*models.Project, error) {
	return a.dataService.CreateProject(host, token, payload)
}

func (a *App) UpdateProject(host, token string, payload map[string]any) (*models.Project, error) {
	return a.dataService.UpdateProject(host, token, payload)
}

func (a *App) DeleteProject(host, token string) (*models.MessageResponse, error) {
	return a.dataService.DeleteProject(host, token)
}

func (a *App) AddDomain(domain string, projectID string) (models.Domain, error) {
	return a.domainService.AddDomain(domain, projectID)
}

func (a *App) GetDomains() []models.Domain {
	return a.domainService.GetDomains()
}

func (a *App) NewConsole() *models.Console {
	return &models.Console{}
}

func (a *App) NewWsMessage() *models.WsMessage {
	return &models.WsMessage{}
}

func (a *App) ConnectSsh(hostname string, port int, username string, password string) (*models.SshConnectResponse, error) {
	return a.terminalService.ConnectSsh(hostname, port, username, password)
}

func (a *App) StartTerminalWebSocket(hostname string, sessionId string) error {
	return a.terminalService.StartTerminalWebSocket(hostname, sessionId)
}

func (a *App) SendTerminalInput(data string) error {
	return a.terminalService.SendTerminalInput(data)
}

func (a *App) ResizeTerminal(cols, rows int) error {
	return a.terminalService.ResizeTerminal(cols, rows)
}

func (a *App) DisconnectTerminal() error {
	return a.terminalService.DisconnectTerminal()
}

func (a *App) IsTerminalConnected() bool {
	return a.terminalService.IsTerminalConnected()
}
