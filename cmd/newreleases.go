package cmd

import (
	"fmt"
	"strings"

	"github.com/RiskIdent/jelease/pkg/newreleases"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var newReleasesCmd = &cobra.Command{
	Use:     "newreleases",
	Aliases: []string{"nr"},
}

var newReleasesDiffCmd = &cobra.Command{
	Use: "diff",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			colorDiffAdd       = color.New(color.FgGreen)
			colorDiffDiverge   = color.New(color.FgRed)
			colorDiffUntracked = color.New(color.FgHiYellow)
		)

		colorize := func(diff string) string {
			lines := strings.Split(diff, "\n")
			for i, line := range lines {
				switch {
				case strings.HasPrefix(line, "+"):
					lines[i] = colorDiffAdd.Sprint(line)
				case strings.HasPrefix(line, "!"):
					lines[i] = colorDiffDiverge.Sprint(line)
				case strings.HasPrefix(line, "?"):
					lines[i] = colorDiffUntracked.Sprint(line)
				}
			}
			return strings.Join(lines, "\n")
		}

		client := newreleases.FromCfg(cfg.NewReleases)
		diff, err := client.Diff()
		if err != nil {
			fmt.Printf("Error while diffing: %q", err)

		}
		fmt.Println(colorize(diff.Summary()))
	},
}

var newReleasesDiffDivergedCmd = &cobra.Command{
	Use: "diverged",
	Run: func(cmd *cobra.Command, args []string) {
		client := newreleases.FromCfg(cfg.NewReleases)
		diff, err := client.Diff()
		if err != nil {
			fmt.Printf("Error while diffing: %q", err)

		}
		fmt.Println(diff.DescribeDiverged())
	},
}

var newReleasesImportCmd = &cobra.Command{
	Use: "import",
	Run: func(cmd *cobra.Command, args []string) {
		nr := newreleases.FromCfg(cfg.NewReleases)
		_, err := nr.ImportProjects()
		if err != nil {
			fmt.Printf("error when importing projects: %v", err)
		}
	},
}

func init() {
	newReleasesDiffCmd.AddCommand(newReleasesDiffDivergedCmd)
	newReleasesCmd.AddCommand(newReleasesDiffCmd)
	newReleasesCmd.AddCommand(newReleasesImportCmd)
	rootCmd.AddCommand(newReleasesCmd)
}
