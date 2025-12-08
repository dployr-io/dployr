// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package scripts

import _ "embed"

//go:embed deploy_app.sh
var DeployScript string

//go:embed systemd.sh
var SystemdScript string

//go:embed system_doctor.sh
var SystemDoctorScript string

//go:embed install.sh
var InstallScript string
