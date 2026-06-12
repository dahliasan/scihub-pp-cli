package cli

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"scihub-pp-cli/internal/cliutil"
	"scihub-pp-cli/internal/scihub"
)

type mirrorProbeResult struct {
	Mirror   string `json:"mirror"`
	Status   string `json:"status"`
	Code     int    `json:"code,omitempty"`
	Latency  string `json:"latency"`
	LatencyMs int64 `json:"latency_ms"`
	Error    string `json:"error,omitempty"`
}

func newProbeCmd(flags *rootFlags) *cobra.Command {
	var mirrorOverride string

	cmd := &cobra.Command{
		Use:   "probe",
		Short: "Probe Sci-Hub mirrors and report the fastest working one",
		Example: `  scihub-pp-cli probe
  scihub-pp-cli probe --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			httpClient := scihub.NewHTTPClient(flags.timeout)

			mirrors, err := scihub.ListMirrors(ctx)
			if err != nil {
				return fmt.Errorf("listing mirrors: %w", err)
			}
			if mirrorOverride != "" {
				mirrors = append([]string{mirrorOverride}, mirrors...)
			}
			if len(mirrors) == 0 {
				return fmt.Errorf("no mirrors to probe")
			}

			seen := make(map[string]struct{})
			var results []mirrorProbeResult
			for _, mirror := range mirrors {
				if _, ok := seen[mirror]; ok {
					continue
				}
				seen[mirror] = struct{}{}
				start := time.Now()
				status, code, probeErr := cliutil.ProbeReachable(ctx, httpClient, mirror+"/")
				elapsed := time.Since(start)
				result := mirrorProbeResult{
					Mirror:    mirror,
					Status:    status,
					Code:      code,
					Latency:   elapsed.Round(time.Millisecond).String(),
					LatencyMs: elapsed.Milliseconds(),
				}
				if probeErr != nil {
					result.Error = probeErr.Error()
				}
				results = append(results, result)
			}

			sort.SliceStable(results, func(i, j int) bool {
				iOK := results[i].Status == cliutil.ReachabilityReachable
				jOK := results[j].Status == cliutil.ReachabilityReachable
				if iOK != jOK {
					return iOK
				}
				return results[i].LatencyMs < results[j].LatencyMs
			})

			var fastest *mirrorProbeResult
			for i := range results {
				if results[i].Status == cliutil.ReachabilityReachable {
					fastest = &results[i]
					break
				}
			}

			if flags.asJSON {
				payload := map[string]any{
					"results": results,
				}
				if fastest != nil {
					payload["fastest"] = fastest.Mirror
				}
				return flags.printJSON(cmd, payload)
			}

			for _, result := range results {
				line := fmt.Sprintf("%s\t%s\t%s", result.Mirror, result.Status, result.Latency)
				if result.Error != "" {
					line += "\t" + result.Error
				}
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			if fastest == nil {
				return fmt.Errorf("no working mirrors found")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "fastest: %s (%s)\n", fastest.Mirror, fastest.Latency)
			return nil
		},
	}
	cmd.Flags().StringVar(&mirrorOverride, "mirror", "", "Additional mirror to include in probe")
	return cmd
}
