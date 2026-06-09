// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/dployr-io/dployr/internal/cli/commands"
)

func main() {
	if err := commands.New().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
