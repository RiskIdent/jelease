// SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/RiskIdent/jelease/pkg/version"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfg config.Config
)

var rootCmd = &cobra.Command{
	Use:           "jelease",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return configSetup()
	},
}

func Execute(defaultConfig config.Config) {
	cfg = defaultConfig

	// Add flag definitons here that need to be binded with configs
	// NOTE: These need to be added AFTER the "cfg = defaultConfig"
	rootCmd.PersistentFlags().String("jira.url", cfg.Jira.URL, "Full URL, including protocol, of the Jira website")
	rootCmd.PersistentFlags().Bool("jira.skipCertVerify", cfg.Jira.SkipCertVerify, "(INSECURE) Skip TLS/SSL certificate verification")
	rootCmd.PersistentFlags().String("jira.auth.token", cfg.Jira.Auth.Token, "Jira personal access token (PAT)")
	rootCmd.PersistentFlags().String("jira.issue.status", cfg.Jira.Issue.Status, "Jira issue status on created issues")
	rootCmd.PersistentFlags().String("jira.issue.type", cfg.Jira.Issue.Type, "Jira issue type on created issues")
	rootCmd.PersistentFlags().String("jira.issue.project", cfg.Jira.Issue.Project, `Jira project name to search for issues in (example: "OP")`)
	rootCmd.PersistentFlags().Uint16("http.port", cfg.HTTP.Port, "Which HTTP port to run the server on.")
	rootCmd.PersistentFlags().String("github.tempdir", util.Deref(cfg.GitHub.TempDir, os.TempDir()), "Which folder to clone repositories into")
	rootCmd.PersistentFlags().Bool("dryrun", cfg.DryRun, "Do not alter any state, e.g skip creating Jira tickets or GitHub PRs")
	rootCmd.PersistentFlags().Var(&cfg.Log.Level, "log.level", "Sets the logging level")
	rootCmd.PersistentFlags().Var(&cfg.Log.Format, "log.format", "Sets the logging format")
	viper.BindPFlags(rootCmd.PersistentFlags())

	loggerSetup() // set up logging first using default config

	err := rootCmd.Execute()
	if err != nil {
		log.Error().Msgf("Failed: %s", err)
		os.Exit(1)
	}
}

func init() {
	// config.FuncsMap must be set before the first time the config is parsed,
	// which happens first in the main() function in main.go (in repo root)

	var (
		pathCharRegex        = regexp.MustCompile(`[^a-zA-Z0-9/,._-]`)
		pathSegmentCharRegex = regexp.MustCompile(`[^a-zA-Z0-9,._-]`)
	)

	config.FuncsMap = template.FuncMap{
		"versionBump": func(add string, to string) string {
			// "1.2.3" | versionBump "0.1.0"
			// add = 0.1.0
			// to = 1.2.3
			addVer, err := version.Parse(add)
			if err != nil {
				panic(fmt.Sprintf("parse version %q: %s", add, err))
			}
			toVer, err := version.Parse(to)
			if err != nil {
				panic(fmt.Sprintf("parse version %q: %s", to, err))
			}
			return toVer.Add(addVer).String()
		},
		"sanitizePath": func(path string) string {
			path = strings.ToLower(path)
			path = filepath.ToSlash(path)
			return pathCharRegex.ReplaceAllLiteralString(path, "-")
		},
		"sanitizePathSegment": func(path string) string {
			path = strings.ToLower(path)
			return pathSegmentCharRegex.ReplaceAllLiteralString(path, "-")
		},
		"basename": func(path string) string {
			return filepath.Base(path)
		},
		"dirname": func(path string) string {
			return filepath.Dir(path)
		},
		"print": func(args ...any) string {
			return fmt.Sprint(args...)
		},
		"printf": func(format string, args ...any) string {
			// "a" | printf "%s-%s" "b"
			// => "b-a"
			return fmt.Sprintf(format, args...)
		},
	}
}

func configSetup() error {
	viper.SetConfigName("jelease")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/jelease/")
	if homePath, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(homePath, ".jelease.yaml"))
	}
	if cfgPath, err := os.UserConfigDir(); err == nil {
		viper.AddConfigPath(cfgPath)
	}
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.TextUnmarshallerHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(), // default hook
		mapstructure.StringToSliceHookFunc(","),     // default hook
	))); err != nil {
		log.Error().Msgf("Failed decoding config file:\n%s", err)
		os.Exit(1)
	}

	// Set up logger again, now that we've read in the new config
	if err := loggerSetup(); err != nil {
		return err
	}

	log.Debug().
		Str("url", cfg.Jira.URL).
		Uint("customField", cfg.Jira.Issue.ProjectNameCustomField).
		Msg("Loaded configuration.")

	return nil
}

func loggerSetup() error {
	pretty := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "Jan-02 15:04",
	})
	switch cfg.Log.Format {
	case config.LogFormatJSON:
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	case config.LogFormatPretty:
		log.Logger = pretty
	default:
		log.Logger = pretty
	}

	log.Logger = log.Level(zerolog.Level(cfg.Log.Level))
	return nil
}
