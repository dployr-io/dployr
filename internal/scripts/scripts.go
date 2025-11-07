package scripts

import _ "embed"

//go:embed deploy_app.sh
var DeployScript string

//go:embed systemd.sh
var SystemdScript string
