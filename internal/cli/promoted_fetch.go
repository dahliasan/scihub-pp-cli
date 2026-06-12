package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"scihub-pp-cli/internal/scihub"
)

func newFetchCmd(flags *rootFlags) *cobra.Command {
	var outputPath string
	var mirrorOverride string

	cmd := &cobra.Command{
		Use:   "fetch <doi>",
		Short: "Download PDF for a DOI via Sci-Hub mirrors",
		Example: `  scihub-pp-cli fetch 10.1038/nature12373 -o paper.pdf
  scihub-pp-cli fetch 10.1038/nature12373 --mirror https://sci-hub.st`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputPath == "" {
				return usageErr(fmt.Errorf("output path is required\nUsage: %s <doi> -o <path>", cmd.CommandPath()))
			}
			if flags.dryRun {
				fmt.Fprintf(cmd.ErrOrStderr(), "dry run: would fetch %s to %s\n", args[0], outputPath)
				return nil
			}
			httpClient := scihub.NewHTTPClient(flags.timeout)
			mirror, err := scihub.FetchDOI(cmd.Context(), httpClient, args[0], outputPath, mirrorOverride)
			if err != nil {
				return fmt.Errorf("fetch failed: %w", err)
			}
			if flags.asJSON {
				return flags.printJSON(cmd, map[string]any{
					"doi":    scihub.NormalizeDOI(args[0]),
					"path":   outputPath,
					"mirror": mirror,
					"status": "ok",
				})
			}
			if flags.quiet {
				fmt.Fprintln(cmd.OutOrStdout(), outputPath)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "saved %s via %s\n", outputPath, mirror)
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output PDF path (required)")
	cmd.Flags().StringVar(&mirrorOverride, "mirror", "", "Sci-Hub mirror base URL override")
	return cmd
}
