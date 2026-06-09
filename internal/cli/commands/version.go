package commands

import (
	"encoding/json"
	"fmt"

	"github.com/dployr-io/dployr/pkg/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Get()
			if jsonOutput {
				data, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else {
				fmt.Println(info.String())
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
	return cmd
}
