package scripts

import _ "embed"

//go:embed setup_runtime.sh
var BashScript string

//go:embed setup_runtime.ps1
var PowershellScript string

//go:embed systemd.sh
var SystemdScript string
