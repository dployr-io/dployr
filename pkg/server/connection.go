package server

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"dployr.io/pkg/logger"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
)

type HostConfig struct {
	User       string
	Host       string
	Port       int
	PrivateKey string
}

type ConnectionPool struct {
	connections map[string]*ssh.Client
	mu          sync.Mutex
}

type CommandResult struct {
	Command  string        `json:"command"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error,omitempty"`
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*ssh.Client),
	}
}

func NewHostConfig(host, username, privateKey string) *HostConfig {
	return &HostConfig{
		User:       username,
		Host:       host,
		Port:       22,
		PrivateKey: privateKey,
	}
}

func (p *ConnectionPool) buildKey(cfg *HostConfig) string {
	return fmt.Sprintf("%s@%s:%d", cfg.User, cfg.Host, cfg.Port)
}

// Get returns an SSH client connection for the given host configuration.
// If a connection already exists in the pool, it returns the existing connection.
// Otherwise, it creates a new SSH connection and stores it in the pool.
func (p *ConnectionPool) Get(cfg *HostConfig) (*ssh.Client, error) {
	hostKeyCallback, err := knownhosts.New("/home/" + cfg.User + "/.ssh/known_hosts")
	if err != nil {
		return nil, fmt.Errorf("unable to load ssh known hosts file: %w", err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	key := p.buildKey(cfg)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Return existing connection if available
	if client, exists := p.connections[key]; exists {
		return client, nil
	}

	// Create new connection
	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
	}

	client, err := ssh.Dial("tcp", cfg.Host+":"+strconv.Itoa(cfg.Port), config)
	if err != nil {
		return nil, err
	}

	p.connections[key] = client
	return client, nil
}

func (p *ConnectionPool) ExecuteCommand(cfg *HostConfig, command string) (*CommandResult, error) {
	client, err := p.Get(cfg)
	if err != nil {
		return nil, err
	}

	// Create session from client
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	start := time.Now()
	err = session.Run(command)
	duration := time.Since(start)

	res := &CommandResult{
		Command:  command,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			res.ExitCode = exitError.ExitStatus()
		} else {
			res.ExitCode = -1
		}
		res.Error = err
	}

	return res, nil
}

func (p *ConnectionPool) BatchExecuteCommand(ctx context.Context, r *repository.EventRepo, cfg *HostConfig, cmds []string, logger *logger.Logger) error {
	for _, cmd := range cmds {
		res, _ := p.ExecuteCommand(cfg, cmd)

		if res.Error != nil {
			logger.Error(ctx, r, models.Setup, res.Stderr)
			return res.Error
		}

		logger.Info(ctx, r, models.Setup, res.Stdout)
	}

	return nil
}
