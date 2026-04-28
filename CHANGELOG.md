## [0.6.3] - 2026-04-28

### 🚀 Features

- *(install)* Enhance Caddy installation with systemd and launchd support
- *(install)* Add Caddy restart logic after directory ownership changes
- *(install)* Enhance caddy service check with restart logic if inactive
## [0.6.2] - 2026-04-28

### 🚀 Features

- *(install)* Add Docker installation logic if not present
- *(install)* Implement run_privileged function for privileged command execution
## [0.6.1] - 2026-04-26

### 🚀 Features

- *(install)* Enhance help message and add install option for dev servers to connect to dev base
- *(install)* Improve error handling for Caddy installation steps
- *(install)* Enhance user creation logic to include docker group if it exists
- *(install)* Improve apt command handling by redirecting input to avoid blocking
- *(install)* Remove default base URL for improved flexibility
- *(auth)* Add logger to ObtainNodeTokenWithBackoff for improved error handling

### ⚙️ Miscellaneous Tasks

- *(formatting)* Gofmt
## [0.6.0] - 2026-04-20

### 🚀 Features

- *(api)* Simplify supported runtime types
- *(core)* Add docker runtime support and service types
- *(install)* Add Docker command permissions to sudoers configuration
- *(install)* Add docker group to dployrd user creation and modification
- *(deploy)* Implement dynamic Dockerfile generation and runtime templates
- *(worker)* Integrate proxy route registration after deployment

### 🐛 Bug Fixes

- *(install)* Improve error logging for instance registration failures

### 🚜 Refactor

- *(web)* Clean comments
- *(core)* Rename agent to node throughout the codebase
- *(core)* Simplify supported runtime types
- *(store)* Remove service_manager field from service model

### 📚 Documentation

- *(service)* Update comments
- *(api)* Update terminology in documentation and schema
- *(api)* Update openapi schema for service types and sources
- *(utils)* Add comment

### 🎨 Styling

- *(install)* Capitalize "Dployr" in service descriptions

### ⚙️ Miscellaneous Tasks

- *(worker)* Simplify deployment log message
## [0.5.11] - 2026-01-18

### 🚜 Refactor

- *(goreleaser)* Fix GitHub raw content URLs
- *(install)* Create Caddy logs directory during installation
## [0.5.10] - 2026-01-18

### 🚀 Features

- *(terminal)* Add interactive terminal support via bi-directional websocket relay from base to client

### 🚜 Refactor

- *(docs)* Remove old store docs
- *(syncer)* Move auth and WebSocket utilities to shared packages
- *(syncer)* Trigger send sync update before executing deploy/proxy tasks instead of after completion
## [0.5.9-beta.1] - 2026-01-15

### 🚜 Refactor

- *(proxy)* Remove HTTPS listener from reverse proxy template to use HTTP only
## [0.5.8] - 2026-01-15

### 🚜 Refactor

- *(install)* Start Caddy service after systemd configuration reload
- *(proxy)* Disable automatic HTTPS in Caddy and explicitly configure HTTP/HTTPS listeners
## [0.5.7] - 2026-01-15

### 🚜 Refactor

- *(install)* Configure Caddy to run as dployrd user with privileged port binding
- *(proxy)* Migrate Caddy management from process to systemd service
## [0.5.6] - 2026-01-15

### 🚜 Refactor

- *(proxy)* Change Caddy default ports from 8080/8443 to 80/443
## [0.5.5] - 2026-01-14

### 🚜 Refactor

- *(system)* Send sync message after deploy task completion
- *(system)* Expand sync trigger to include deployment and proxy tasks
- *(system)* Add Caddy restart to system restart and reboot operations
## [0.5.4] - 2026-01-11

### 🚜 Refactor

- *(system)* Fix workload conversions in update builder
## [0.5.3] - 2026-01-11

### 🚜 Refactor

- *(release)* Update installation script URLs to use correct path
- *(system)* Add error handling to syncer
## [0.5.2] - 2026-01-10

### 🚜 Refactor

- *(api)* Make services a mandatory field in services struct
- *(system)* Stable schema contract for update v1.1
- *(release)* Update CDN for installation script
- *(system)* Cleanups and improvement
## [0.5.1] - 2026-01-06

### 🚜 Refactor

- *(install)* Fix path to Caddyfile in sudoers permissions
## [0.5.0] - 2026-01-06

### 🚜 Refactor

- *(proxy)* Updated unstructured remove-route request
- *(api)* Implement v1.1 update schema
## [0.4.13-beta.26] - 2026-01-02

### 🚜 Refactor

- *(proxy)* Enhance proxy status response with detailed Caddy configuration
## [0.4.13-beta.25] - 2026-01-02

### 🚜 Refactor

- *(api)* Update docs
## [0.4.13-beta.24] - 2026-01-02

### 🚜 Refactor

- *(proxy)* Updated remaining unstructured logs to structured logs in Caddy handler
## [0.4.13-beta.23] - 2026-01-02

### 🚜 Refactor

- *(proxy)* Update proxy creation to detailed payload
- *(proxy)* Updated to structured logs in Caddy handler
## [0.4.13-beta.22] - 2026-01-01

### 🚀 Features

- *(install)* Add automatic git installation to installer
## [0.4.13-beta.21] - 2026-01-01

### 📚 Documentation

- *(readme)* Update documentation URLs from docs.dployr.dev to dployr.io/docs
## [0.4.13-beta.20] - 2025-12-31

### 🚀 Features

- *(ci)* Add automated changelog generation with git-cliff
## [0.4.13-beta.19] - 2025-12-30

### 🚜 Refactor

- *(system)* Add more system metrics
## [0.4.13-beta.18] - 2025-12-29

### 🚜 Refactor

- *(service)* Remove status field and populate blueprint data from deployment
## [0.4.13-beta.17] - 2025-12-29

### 🚜 Refactor

- *(deploy)* Change env vars & secrets to accept any type
## [0.4.13-beta.16] - 2025-12-29

### 🚜 Refactor

- *(api)* Simplify list endpoints to return flat array data
## [0.4.13-beta.15] - 2025-12-28

### 🚜 Refactor

- *(deploy)* Remove timestamp prefix from deployment log messages to prevent duplicates
## [0.4.13-beta.14] - 2025-12-28

### 🚜 Refactor

- *(deploy)* Capture and display service runtime logs on startup failure
## [0.4.13-beta.13] - 2025-12-28

### 🚜 Refactor

- *(logs)* Skip partial line in log stream
## [0.4.13-beta.12] - 2025-12-28

### 🚜 Refactor

- *(logger)* Update deployment log format to write raw messages
- *(syncer)* Disable task deduplication for log streaming tasks
## [0.4.13-beta.11] - 2025-12-28

### 🚜 Refactor

- *(deploy)* Stream deployment script output to log file instead of stdout/stderr
## [0.4.13-beta.10] - 2025-12-28

### 🚜 Refactor

- *(logs)* Improve stream positioning logic with explicit mode handling and fallback behavior
## [0.4.13-beta.9] - 2025-12-28

### 🚜 Refactor

- *(executor)* Stream entire deployment log instead of defaulting to live mode
## [0.4.13-beta.8] - 2025-12-28

### 🚜 Refactor

- *(syncer)* Change ws endpoint parameter from instanceId to instanceName
## [0.4.13-beta.7] - 2025-12-28

### 🚜 Refactor

- *(syncer)* Change agent cert endpoint parameter from instanceId to instanceName
## [0.4.13-beta.6] - 2025-12-24

### 🚜 Refactor

- *(syncer)* Change WebSocket and certificate endpoints to use query parameters instead of path parameters
- *(api)* Remove unused agent certificate endpoints from OpenAPI specification
## [0.4.13-beta.5] - 2025-12-24

### 🚜 Refactor

- *(service)* Change service lookup from ID to name across API, CLI, and handlers
- *(logger)* Change default output from stdout to stderr when rotating writer is unavailable
- *(logs)* Normalize log file paths to lowercase before processing
## [0.4.13-beta.3] - 2025-12-23

### 🚜 Refactor

- *(worker)* Improve semaphore test with proper goroutine synchronization and timeouts
## [0.4.13-beta.2] - 2025-12-23

### 🚀 Features

- *(filesystem)* Add real-time directory watching with fsnotify and WebSocket events
- *(service)* Add systemd service and proxy cleanup on service deletion
- *(config)* Add custom instance ID support to installation scripts and agent configuration

### 🚜 Refactor

- *(executor)* Remove unused request and trace context setup from task execution
- *(syncer)* Sanitize secrets in agent updates by replacing values with empty strings
## [0.4.13-beta.1] - 2025-12-21

### 🚀 Features

- *(filesystem)* Add filesystem browsing and file operations API
- *(system)* Add endpoint for gopsutil-based resource monitoring

### 🚜 Refactor

- *(deploy)* Add duplicate key detection to service env
- *(service)* Add secrets support and config.toml-based environment management
- *(auth)* Add fallback to RegisteredClaims.Subject when claims.Subject is empty
- *(store)* Add description field to services table and use sql.NullString for nullable columns
- *(logs)* Add time-based filtering, file rotation detection, and rate limiting to log streaming
- *(filesystem)* Impove file system implementation and real-time ws response for real-time requests
## [0.4.12] - 2025-12-17

### 🚜 Refactor

- *(logs)* Separate deployment and service log paths, write deployment logs directly to file
- *(deploy)* Simplify CloneRepo by removing workDir parameter and cleaning destDir before clone
- *(shell)* Capture stderr to buffer and include error messages for better debugging
- *(service)* Update domain request to base from /v1/domains to /v1/domains/register to register quick domain
## [0.4.11-beta.1] - 2025-12-16

### 🚀 Features

- *(system)* Extend service counts with full deployment, service, and proxy app lists in system status
- *(logger)* Add log rotation with 300MB size and 24-hour age limits, keeping 5 backups
- *(deploy)* Add custom JSON unmarshaler for DeployRequest runtime field
- *(db)* Enable WAL mode and set 5s busy timeout to prevent SQLITE_BUSY errors

### 🚜 Refactor

- *(deploy)* Simplify limit capping logic using min function
- *(deploy)* Remove unused DNSProvider field from DeployRequest
- *(deploy)* Remove dns-provider flag from CLI deploy command
- *(db)* Set single connection limit and move WAL/busy_timeout to DSN to prevent SQLITE_BUSY errors
- *(deploy)* Add version field to DeployRequest unmarshaler and validate RuntimeObj type before assignment
- *(deploy)* Simplify DeployRequest by flattening Runtime field to string from object
## [0.4.10] - 2025-12-09

### 🚀 Features

- *(system)* Split restart into daemon restart and OS reboot endpoints, enhance install script with auth and task exclusion
## [0.4.9] - 2025-12-08

### 🚀 Features

- *(system)* Add token validation for log streaming, include worker count in system updates
## [0.4.8] - 2025-12-08

### 🚀 Features

- *(system)* Add system restart with systemctl
## [0.4.7] - 2025-12-08

### 🚀 Features

- *(install)* Add automated installation script with graceful daemon upgrade handling
## [0.4.7-beta.6] - 2025-12-04

### 🚜 Refactor

- *(logs)* Replace logType with path for flexible log file routing
## [0.4.7-beta.5] - 2025-12-04

### 🚀 Features

- *(executor)* Pass context to log stream handler and suppress context cancellation errors
## [0.4.7-beta.4] - 2025-12-03

### 🚀 Features

- *(ci)* Enhance Discord release notification with richer embed formatting
- *(install)* Add configurable base URL option with --base flag
- *(install)* Update registration API endpoint to use versioned daemon URL and enhance error handling

### 🐛 Bug Fixes

- *(install)* Fix registration response structure
## [0.4.7-beta.2] - 2025-11-30

### 🚀 Features

- *(logs)* Implement adaptive batching for log streaming with 250ms window and 50-line threshold
## [0.4.7-beta.1] - 2025-11-30

### 🚀 Features

- *(logs)* Add size-based chunking to prevent WebSocket message size limits

### 🚜 Refactor

- *(logs)* Consolidate WebSocket log streaming functionality with existing open connections
## [0.4.6] - 2025-11-30

### 🚀 Features

- *(logs)* Replace SSE log streaming with WebSocket-based implementation

### 🚜 Refactor

- *(license)* Change license from MIT to Apache-2.0 and add copyright headers to all files
## [0.4.6-beta.8] - 2025-11-28

### 🚀 Features

- *(logs)* Add log streaming functionality with WebSocket support
## [0.4.6-beta.7] - 2025-11-28

### 🚜 Refactor

- *(system)* Remove WebSocket message ID from logger context
## [0.4.6-beta.6] - 2025-11-28

### 💼 Other

- *(system)* Update WebSocket message field names to use snake_case and make timestamp required
## [0.4.6-beta.5] - 2025-11-28

### 💼 Other

- *(system)* Fix missing requestId field in WebSocket messages
## [0.4.6-beta.4] - 2025-11-28

### 💼 Other

- *(system)* Add per-message and per-task context tracing with unique IDs
## [0.4.6-beta.3] - 2025-11-27

### 🚜 Refactor

- *(system)* Add bootstrap token to auth debug info and simplify WebSocket schema names
## [0.4.6-beta.2] - 2025-11-27

### 🚜 Refactor

- *(system)* Remove Go version from platform info and fix component name in build flags
## [0.4.4] - 2025-11-27

### 🚜 Refactor

- *(system)* Update version entry to buildinfo on agent update
## [0.4.3] - 2025-11-27

### 🚜 Refactor

- *(utils)* Fix free memory to show correct values
## [0.4.3-beta.2] - 2025-11-26

### 🚜 Refactor

- *(ci)* Configure Go proxy settings for module publishing
## [0.4.3-beta.1] - 2025-11-26

### 🚀 Features

- *(system)* Add comprehensive system resource metrics and enhanced status reporting
## [0.4.2] - 2025-11-26

### 🚜 Refactor

- *(system)* Move WebSocket message schemas to pkg/core/system
## [0.4.1] - 2025-11-26

### 🚜 Refactor

- *(ci)* Simplify release workflow conditions
- *(release)* Improve tag handling and beta-to-stable promotion
## [0.4.1-beta.14] - 2025-11-26

### 📚 Documentation

- *(readme)* Update overview section and deployment examples
## [0.4.1-beta.13] - 2025-11-26

### 🚜 Refactor

- *(syncer)* Add periodic agent status updates and improve WebSocket message handling

### 📚 Documentation

- *(contributing)* Clarify documentation requirement
## [0.4.1-beta.12] - 2025-11-25

### 🚜 Refactor

- *(system)* Bug fixes on version lookup, and improve WebSocket error handling
## [0.4.1-beta.11] - 2025-11-25

### 🚀 Features

- *(metrics)* Add Prometheus metrics endpoint with task execution and system health tracking
## [0.4.1-beta.10] - 2025-11-25

### 🚜 Refactor

- *(executor)* Add AccessTokenProvider interface and inject token authentication into task execution
## [0.4.1-beta.9] - 2025-11-25

### 🐛 Bug Fixes

- *(store)* Parse timestamp columns as strings to handle SQLite TEXT storage
## [0.4.1-beta.8] - 2025-11-25

### 🐛 Bug Fixes

- *(install)* Add fallback to static jq binary download when apt install fails

### 🚜 Refactor

- *(db)* Reorder instance table columns to group token fields together
## [0.4.1-beta.7] - 2025-11-25

### 🐛 Bug Fixes

- *(syncer)* Change default client cert directory from /etc/dployr to /var/lib/dployrd on Linux

### 🚜 Refactor

- *(module)* Migrate from local to GitHub module path github.com/dployr-io/dployr
## [0.4.1-beta.6] - 2025-11-25

### 🐛 Bug Fixes

- *(tests)* Fixed race detection and add thread-safety to mock deployment store

### 🚜 Refactor

- *(logging)* Replace slog.Logger with shared.Logger and add structured logging to syncer and executor

### 📚 Documentation

- *(readme)* Add comprehensive documentation with badges, quickstart guide, and troubleshooting

### 🎨 Styling

- Fix whitespace and add missing newlines at end of files

### ⚙️ Miscellaneous Tasks

- Intoduced local ci parity of ci handlers for quick local development
## [0.4.1-beta.5] - 2025-11-24

### 🐛 Bug Fixes

- *(store)* Handle db NULL values for bootstrap_token and access_token
## [0.4.1-beta.4] - 2025-11-24

### 🚜 Refactor

- *(store)* Split token into bootstrap_token and access_token fields
## [0.4.1-beta.3] - 2025-11-24

### 🚀 Features

- *(agent)* Add WebSocket-based task syncing with mTLS client authentication
## [0.4.1-beta.2] - 2025-11-23

### 🚀 Features

- *(system)* Add bootstrap token rotation and registration status endpoints
## [0.4.1-beta.1] - 2025-11-23

### 🐛 Bug Fixes

- *(version)* Add bounds checking for commit hash truncation
## [0.4.0] - 2025-11-23

### 🚀 Features

- *(system)* Add task syncing and daemon mode management
## [0.3.1-beta.24] - 2025-11-23

### 🐛 Bug Fixes

- *(install)* Add apt lock handling and allow initial instance registration updates
## [0.3.1-beta.23] - 2025-11-22

### 🚜 Refactor

- *(db)* Bug fix on instance table immutability constraints and simplify trigger logic
## [0.3.1-beta.22] - 2025-11-22

### 🚀 Features

- *(store)* Add fallback for SetToken store method for fresh installs
## [0.3.1-beta.21] - 2025-11-22

### 🚜 Refactor

- *(auth)* Update JWKS endpoint path and remove redundant config field
## [0.3.1-beta.20] - 2025-11-22

### 🚜 Refactor

- *(logging)* Simplify context value extraction in LogWithContext
## [0.3.1-beta.19] - 2025-11-22

### 🚀 Features

- *(logging)* Add structured logging to handlers
## [0.3.1-beta.18] - 2025-11-22

### 🚜 Refactor

- *(auth)* Initialize Auth service with InstanceStore dependency
## [0.3.1-beta.17] - 2025-11-22

### 🚜 Refactor

- *(api)* Standardize error codes
## [0.3.1-beta.16] - 2025-11-22

### 🚜 Refactor

- *(system)* Extract InstallRequest into reusable schema
## [0.3.1-beta.15] - 2025-11-22

### 🐛 Bug Fixes

- *(system)* Remove unused address field from domain registration request
## [0.3.1-beta.14] - 2025-11-22

### 🚀 Features

- *(install)* Display registered domain in installation summary
- *(install)* Separate structured logs from user output using dedicated file descriptors
- *(install)* Improve domain registration error handling
- *(install)* Add specific error handling for used tokens during registration
- *(install)* Enhance error messages with detailed error information during registration

### 🐛 Bug Fixes

- *(install)* Update domain extraction path from registration response
## [0.3.1-beta.13] - 2025-11-22

### 🚀 Features

- *(api)* Add response schema for domain registration endpoint
- *(install)* Refactor installer with structured logging and improved error handling. Implemented quiet install for system updates
## [0.3.1-beta.12] - 2025-11-22

### 🚀 Features

- *(install)* Improve error handling and output for installer

### 🐛 Bug Fixes

- *(system)* Add nil check for instance before accessing InstanceID
## [0.3.1-beta.11] - 2025-11-22

### 🚀 Features

- *(install)* Update domain registration payload key from 'claim' to 'token'

### 🐛 Bug Fixes

- *(store)* Rename database table from 'instances' to 'instance'
## [0.3.1-beta.10] - 2025-11-22

### 🚀 Features

- *(system)* Update store instance assignment from install
- *(system)* Add public IP support and domain registration endpoint
- *(install)* Add error log output to stderr for better debugging
- *(install)* Add real-time log output to installer
- *(http)* Handle errors conversion properly

### 📚 Documentation

- *(install)* Update installation documentation
- *(install)* Update help examples and remove duplicate help flag handling

### 🧪 Testing

- *(deploy)* Remove unused buildAuthUrl test cases
## [0.3.1-beta.8] - 2025-11-21

### 🚀 Features

- *(system)* Update domain registration endpoint path to match base API
## [0.3.1-beta.7] - 2025-11-21

### 🚀 Features

- *(install)* Add required install token parameter for instance registration at installation
- *(system)* Improved token comparison to use constant-time for base64-encoded tokens in instance registration
## [0.3.1-beta.6] - 2025-11-21

### 🚀 Features

- *(system)* Add instance registration and domain management with base integration
## [0.3.1-beta.5] - 2025-11-20

### 🚀 Features

- *(system)* Add system doctor and install endpoints with enhanced system info for dependency checks (vfox, caddy, data directory)
- *(install)* Add installation logging and update config for base authentication
- *(api)* Standardize error handling and improve response formats across handlers

### 📚 Documentation

- *(api)* Simplify authentication documentation in OpenAPI spec
## [0.3.1-beta.3] - 2025-11-19

### 📚 Documentation

- *(api)* Add authentication documentation link to OpenAPI spec
## [0.3.1-beta.2] - 2025-11-19

### 🚀 Features

- *(docs)* Simplify authentication flow documentation in OpenAPI spec
## [0.3.1-beta.1] - 2025-11-19

### 🚀 Features

- *(auth)* Remove authentication endpoints
## [0.2.2-beta.5] - 2025-11-18

### 🚀 Features

- *(auth)* Migrate from HMAC to RSA-256 JWT signing with bootstrap tokens
## [0.2.2-beta.2] - 2025-11-17

### 🚀 Features

- *(api)* Update log streaming endpoint to use Server-Sent Events (SSE)

### 📚 Documentation

- *(api)* Update OpenAPI log streaming documentation

### ⚙️ Miscellaneous Tasks

- *(openapi)* Repostitory dispatch to autogenerate SDK from openapi spec using Kiota
- *(docs)* Update description
- *(docs)* Update description
- *(release)* Add Discord notification on release publication
## [0.2.1] - 2025-11-07

### 🐛 Bug Fixes

- *(install)* Improve Caddy installation process and apt lock handling

### 🧪 Testing

- *(unit)* Add test suite

### ⚙️ Miscellaneous Tasks

- Update installer to handle steps one at a time
## [0.2.1-beta.11] - 2025-11-07

### 🐛 Bug Fixes

- *(deploy)* Improve systemd service deployment and runtime environment handling

### ⚙️ Miscellaneous Tasks

- Gofmt
## [0.2.1-beta.10] - 2025-11-07

### 🚀 Features

- *(install)* Enhance system configuration and environment setup
- *(runtime)* Enhance cross-platform runtime setup with vfox scripts
- *(deploy)* Implement unified deployment script and runtime setup

### 🐛 Bug Fixes

- *(install)* Improve binary installation error handling

### ⚙️ Miscellaneous Tasks

- Cleanup
## [0.2.1-beta.9] - 2025-11-06

### 🐛 Bug Fixes

- *(release)* Improve version retrieval in release script

### 🚜 Refactor

- *(deploy)* Simplify runtime and dependency setup process
## [0.2.1-beta.5] - 2025-11-05

### 🐛 Bug Fixes

- *(auth)* Improve unauthorized token error handling
## [0.2.1-beta.4] - 2025-11-05

### 🐛 Bug Fixes

- *(install)* Reload shell after installing vfox plugins
## [0.2.1-beta.3] - 2025-11-05

### 🐛 Bug Fixes

- *(deploy)* Improve runtime and dependency installation process
## [0.2.1-beta.2] - 2025-11-05

### 🐛 Bug Fixes

- *(deploy)* Simplify runtime installation process
## [0.2.1-beta.1] - 2025-11-05

### 🐛 Bug Fixes

- *(auth)* Improve token management and error handling
## [0.2.0] - 2025-11-05

### 🚀 Features

- *(install)* Add vfox plugin setup for multiple runtimes
- *(install)* Add role-based system groups for dployrd
- *(auth)* Implement role-based user management and access control
- *(auth)* Enhance authentication and authorization system

### ⚙️ Miscellaneous Tasks

- *(goreleaser)* Update version component naming for CLI and daemon
## [0.1.1-beta.33] - 2025-11-05

### 🐛 Bug Fixes

- *(deploy)* Improve runtime setup with correct working directory
## [0.1.1-beta.32] - 2025-11-05

### 📚 Documentation

- *(readme)* Update README with minor text and branch reference adjustments
## [0.1.1-beta.30] - 2025-11-05

### 🐛 Bug Fixes

- *(auth)* Remove default token expiry fallback in CLI
## [0.1.1-beta.29] - 2025-11-05

### 🚀 Features

- *(install)* Add vfox version manager to installation process
- *(cli)* Add logs command for viewing deployment logs
## [0.1.1-beta.28] - 2025-11-05

### 🚜 Refactor

- *(utils)* Centralize data directory path handling across application
## [0.1.1-beta.27] - 2025-11-05

### 🚀 Features

- *(db)* Improve database path handling for cross-platform compatibility
## [0.1.1-beta.26] - 2025-11-05

### 🚀 Features

- *(install)* Improve daemon management during installation
- *(install)* Improve service management and user isolation for dployrd
## [0.1.1-beta.25] - 2025-11-05

### 🐛 Bug Fixes

- *(auth)* Improve token generation with better error handling and default lifespan
- *(version)* Improve version string formatting for git commit hash

### 📚 Documentation

- *(readme)* Update deployment example with real example
## [0.1.1-beta.24] - 2025-11-05

### 🚀 Features

- *(cli)* Enhance dployr CLI description and documentation
- *(runtime)* Standardize runtime type from "node-js" to "nodejs"
## [0.1.1-beta.23] - 2025-11-05

### 🚀 Features

- *(install)* Improve installation and service management for dployr
## [0.1.1-beta.22] - 2025-11-05

### 🚀 Features

- *(install)* Add version selection and help options for installers
- *(config)* Update system-wide configuration location and handling
## [0.1.1-beta.21] - 2025-11-05

### 🚀 Features

- *(install)* Add default configuration generation for dployr

### 📚 Documentation

- *(install)* Update installation instructions and scripts
## [0.1.1-beta.20] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Simplify release workflow by removing Chocolatey upload step for now
## [0.1.1-beta.19] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update Chocolatey package push URL for release workflow
## [0.1.1-beta.18] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update Chocolatey package push path for release workflow
## [0.1.1-beta.17] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Fix Chocolatey publish step shell configuration
- *(workflows)* Refactor Chocolatey package generation and publishing step
## [0.1.1-beta.16] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update Chocolatey package push to use bash as shell
## [0.1.1-beta.15] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Simplify Chocolatey publish step in release workflow
## [0.1.1-beta.14] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(goreleaser)* Remove Chocolatey publish flag
## [0.1.1-beta.13] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Refactor release workflow for improved artifact handling and Chocolatey deployment
## [0.1.1-beta.12] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update GoReleaser v2 release arguments for release workflow
## [0.1.1-beta.11] - 2025-11-03

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update GoReleaser release arguments for Chocolatey-only deployment
## [0.1.1-beta.10] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(goreleaser)* Update archive configuration
## [0.1.1-beta.9] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(workflows)* Typo
## [0.1.1-beta.7] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update GoReleaser v2 release arguments
## [0.1.1-beta.6] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(workflows)* Typo fix
## [0.1.1-beta.5] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(goreleaser)* Add Chocolatey owners and API key configuration
- *(workflows)* Remove beta workflow file
## [0.1.1-beta.4] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update GoReleaser v2 arguments for beta and release workflows
## [0.1.1-beta.3] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update GoReleaser v2 skip parameters for beta and release workflows
## [0.1.1-beta.2] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(workflows)* Update to GoReleaser v2 argument conventions
## [0.1.1-beta.1] - 2025-11-02

### ⚙️ Miscellaneous Tasks

- *(workflows)* Add Chocolatey version flag to GitHub Actions
## [0.1.0] - 2025-11-02

### 🚀 Features

- Add initial website structure with HTML, CSS, and JavaScript for dployr.io
- Update on client application with new features and UI improvements

### 💼 Other

- Implementation for projects deployment
- Refresh tokens implementation
- Bug fixes and improvements in the service creation flow
- Updated deployments page with table

### 📚 Documentation

- *(project)* Add comprehensive contributing and readme documentation

### ⚙️ Miscellaneous Tasks

- Init version 0.1.0 + devbump
- Cleanup
- Go.work.sum
- *(workflows)* Refactor release and beta workflows for Go project
- *(workflows)* Split release and beta workflows into two-stage jobs
- *(workflows)* Fix GoReleaser Chocolatey skip parameter
- *(workflows)* Title fix
- *(workflows)* Update GitHub token for release and beta workflows
- *(workflows)* Update title
- *(goreleaser)* Override GitHub tokens for Homebrew and Scoop tap repositories
- *(goreleaser)* Remove hardcoded GitHub tokens from tap repositories
- *(workflows)* Typo
