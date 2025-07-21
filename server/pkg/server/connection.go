package server

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
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

