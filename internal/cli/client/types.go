package client

// API response types mirror the dployr-base TypeScript interfaces.

type Cluster struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Owner     string   `json:"owner,omitempty"`
	Role      string   `json:"role,omitempty"` // owner | admin | developer | viewer
	CreatedAt UnixTime `json:"createdAt"`
	UpdatedAt UnixTime `json:"updatedAt"`
}

type ClusterUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"` // owner | admin | developer | viewer
}

type Service struct {
	ID             string    `json:"id"`
	ClusterID      string    `json:"clusterId"`
	Name           string    `json:"name"`
	Label          string    `json:"label"`
	Type           string    `json:"type"` // static | web | worker | job
	DeploymentID   *string   `json:"deploymentId,omitempty"`
	DeploymentName *string   `json:"deploymentName,omitempty"`
	IcedAt         *UnixTime `json:"icedAt,omitempty"`
	CreatedAt      UnixTime  `json:"createdAt"`
	UpdatedAt      UnixTime  `json:"updatedAt"`
}

type Deployment struct {
	ID               string    `json:"id"`
	ClusterID        string    `json:"clusterId"`
	ServiceID        string    `json:"serviceId"`
	UserID           string    `json:"userId"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	Source           string    `json:"source"` // remote | image
	Status           string    `json:"status"` // pending | running | success | failed
	Description      string    `json:"description"`
	RunCmd           string    `json:"runCmd"`
	BuildCmd         string    `json:"buildCmd"`
	Port             int       `json:"port"`
	WorkingDir       string    `json:"workingDir"`
	StaticDir        string    `json:"staticDir"`
	Image            string    `json:"image"`
	Domain           string    `json:"domain"`
	RuntimeType      string    `json:"runtimeType"`
	RuntimeVersion   string    `json:"runtimeVersion"`
	RemoteURL        string    `json:"remoteUrl"`
	RemoteBranch     string    `json:"remoteBranch"`
	RemoteCommitHash string    `json:"remoteCommitHash"`
	CreatedAt        UnixTime  `json:"createdAt"`
	UpdatedAt        UnixTime  `json:"updatedAt"`
	FinishedAt       *UnixTime `json:"finishedAt,omitempty"`
}

type CreateDeploymentResult struct {
	TaskID string `json:"taskId"`
	Cached bool   `json:"cached,omitempty"`
}

type CreateDeploymentRequest struct {
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	Source           string            `json:"source"` // remote | image
	Type             string            `json:"type,omitempty"`
	RuntimeType      string            `json:"runtimeType"`
	RuntimeVersion   string            `json:"runtimeVersion,omitempty"`
	RunCmd           string            `json:"runCmd,omitempty"`
	BuildCmd         string            `json:"buildCmd,omitempty"`
	Port             int               `json:"port,omitempty"`
	WorkingDir       string            `json:"workingDir,omitempty"`
	StaticDir        string            `json:"staticDir,omitempty"`
	HealthCheck      string            `json:"healthCheck,omitempty"`
	Image            string            `json:"image,omitempty"`
	Domain           string            `json:"domain,omitempty"`
	RemoteURL        string            `json:"remoteUrl,omitempty"`
	RemoteBranch     string            `json:"remoteBranch,omitempty"`
	RemoteCommitHash string            `json:"remoteCommitHash,omitempty"`
	EnvVars          map[string]string `json:"envVars,omitempty"`
	Secrets          map[string]string `json:"secrets,omitempty"`
	ForceRebuild     bool              `json:"forceRebuild,omitempty"`
}

type Instance struct {
	ID        string   `json:"id"`
	Kind      string   `json:"kind"` // dedicated | pool
	ClusterID *string  `json:"clusterId,omitempty"`
	Address   string   `json:"address"`
	Tag       string   `json:"tag"`
	Status    string   `json:"status"` // healthy | degraded | offline | unreachable | maintenance | provisioning
	Role      string   `json:"role"`   // instance | build
	Managed   bool     `json:"managed"`
	Region    string   `json:"region,omitempty"`
	CreatedAt UnixTime `json:"createdAt"`
	UpdatedAt UnixTime `json:"updatedAt"`
}

type Me struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture,omitempty"`
	Provider string `json:"provider"`
}

// PendingInvite is returned by GET /v1/clusters/users/invites.
type PendingInvite struct {
	ClusterID   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	OwnerName   string `json:"ownerName"`
}

// LoginEmailRequest is the body for POST /v1/auth/login/email.
type LoginEmailRequest struct {
	Email string `json:"email"`
}

// VerifyEmailRequest is the body for POST /v1/auth/login/email/verify.
type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// LogChunk is a single log entry returned from the log API.
type LogChunk struct {
	Timestamp UnixTime `json:"timestamp"`
	Source    string   `json:"source"` // build | runtime
	Level     string   `json:"level"`  // info | warn | error
	Message   string   `json:"message"`
}

// ApiToken is a scoped personal access token (dpat_).
type ApiToken struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Scopes     []string  `json:"scopes"`
	CreatedAt  UnixTime  `json:"createdAt"`
	ExpiresAt  *UnixTime `json:"expiresAt,omitempty"`
	LastUsedAt *UnixTime `json:"lastUsedAt,omitempty"`
}

// CreatedApiToken is returned once on token creation — includes the plaintext.
type CreatedApiToken struct {
	ApiToken
	Token string `json:"token"`
}

// LogSubscribeRequest is sent over WebSocket to subscribe to a service log stream.
type LogSubscribeRequest struct {
	Kind      string `json:"kind"` // "log_subscribe"
	ServiceID string `json:"serviceId"`
	Source    string `json:"source"` // build | runtime | all
}
