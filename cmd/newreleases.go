package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/RiskIdent/jelease/pkg/newreleases"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var newReleasesCmd = &cobra.Command{
	Use:     "newreleases",
	Short:   "Manager your newreleases.io account",
	Aliases: []string{"nr"},
}

var newReleasesDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show the differences between your local configuration and the configured newreleases.io account",
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
	Use:   "diverged",
	Short: "Show details about the diverged project configurations",
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
	Use:   "import",
	Short: "Imports existing resources on newreleases.io and outputs as configuration to add to your config file",
	Run: func(cmd *cobra.Command, args []string) {
		nr := newreleases.FromCfg(cfg.NewReleases)
		diff, err := nr.Diff()
		if err != nil {
			fmt.Printf("Error while diffing: %q", err)
		}
		missingOnLocalSlice := newreleases.SliceFromMap(diff.MissingOnLocal)
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		defer enc.Close()
		if err = enc.Encode(missingOnLocalSlice); err != nil {
			fmt.Printf("Error when encoding import from remote %s", err)
		}
	},
}

var newReleasesApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Applies your local configuration to the configured newreleases.io account",
	Run: func(cmd *cobra.Command, args []string) {
		nr := newreleases.FromCfg(cfg.NewReleases)
		applyOptions := newreleases.ApplyLocalConfigOptions{}

		err := nr.ApplyLocalConfig(applyOptions)
		if err != nil {
			fmt.Printf("Error when applying %s", err)
		}

	},
}

// TODO:
// implement equal for newreleases.project instead, make it ignore ID stuff
// when converting from ProjectCfg, take some globally configured fields

func init() {
	newReleasesDiffCmd.AddCommand(newReleasesDiffDivergedCmd)
	newReleasesCmd.AddCommand(newReleasesDiffCmd)
	newReleasesCmd.AddCommand(newReleasesImportCmd)
	newReleasesCmd.AddCommand(newReleasesApplyCmd)
	rootCmd.AddCommand(newReleasesCmd)
}
