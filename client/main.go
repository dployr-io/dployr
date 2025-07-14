package main

import (
	"context"
	"dployr/core/auth"
	"dployr/core/data"
	"dployr/core/domain"
	"dployr/core/http"
	"dployr/core/terminal"
	"dployr/core/types"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS


type App struct {
	ctx             context.Context
	authService     *auth.AuthService
	httpClient      *http.Client
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
		Title:  "dployr desktop",
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
	baseURL := getBaseUrl()
	
	a.httpClient = http.NewClient()
	a.authService = auth.NewAuthService(baseURL)
	a.dataService = data.NewDataService()
	a.domainService = domain.NewDomainService(a.httpClient, baseURL)
	a.terminalService = terminal.NewTerminalService(ctx, a.httpClient)
}

func getBaseUrl() string {
	return "";
}

// Wails binding methods - delegate to services
func (a *App) SignIn(provider string) types.AuthResponse {
	return a.authService.SignIn(provider)
}

func (a *App) SignOut() bool {
	return a.authService.SignOut()
}

func (a *App) GetCurrentUser() *types.User {
	return a.authService.GetCurrentUser()
}

func (a *App) StoreSession(token string) {
	a.authService.StoreSession(token)
}

func (a *App) FetchData(url string) (any, error) {
	return a.httpClient.Get(url)
}

func (a *App) PostData(url string, data interface{}) (any, error) {
	return a.httpClient.Post(url, data)
}

func (a *App) UpdateData(url string, data interface{}) (any, error) {
	return a.httpClient.Put(url, data)
}

func (a *App) DeleteData(url string) (any, error) {
	return a.httpClient.Delete(url)
}

func (a *App) GetDeployments() []types.Deployment {
	return a.dataService.GetDeployments()
}

func (a *App) GetLogs() []types.LogEntry {
	return a.dataService.GetLogs()
}

func (a *App) GetProjects() []types.Project {
	return a.dataService.GetProjects()
}

func (a *App) AddDomain(domain string, projectID string) (types.Domain, error) {
	return a.domainService.AddDomain(domain, projectID)
}

func (a *App) GetDomains() []types.Domain {
	return a.domainService.GetDomains()
}

func (a *App) NewConsole() *types.Console {
	return &types.Console{}
}

func (a *App) NewWsMessage() *types.WsMessage {
	return &types.WsMessage{}
}

func (a *App) ConnectSsh(hostname string, port int, username string, password string) (*types.SshConnectResponse, error) {
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


