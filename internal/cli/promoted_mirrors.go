// Copyright 2026 dahlia. Licensed under Apache-2.0. See LICENSE.
// Hand-authored workflow command replacing generated REST mirror list.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"scihub-pp-cli/internal/scihub"
)

func newMirrorsPromotedCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mirrors",
		Short:   "List active Sci-Hub mirror URLs from sci-hub.pub",
		Long:    "Fetch https://www.sci-hub.pub/ and parse discovered Sci-Hub mirror links.",
		Example: "  scihub-pp-cli mirrors\n  scihub-pp-cli mirrors --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.dryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "dry run: would fetch mirror list from sci-hub.pub")
				return nil
			}
			mirrors, err := scihub.ListMirrors(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing mirrors: %w", err)
			}
			if flags.asJSON {
				return flags.printJSON(cmd, map[string]any{
					"mirrors": mirrors,
					"count":   len(mirrors),
				})
			}
			for _, mirror := range mirrors {
				fmt.Fprintln(cmd.OutOrStdout(), mirror)
			}
			return nil
		},
	}
	return cmd
}
