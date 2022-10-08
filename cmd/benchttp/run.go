package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/benchttp/engine/runner"

	"github.com/benchttp/cli/internal/configfile"
	"github.com/benchttp/cli/internal/configflag"
	"github.com/benchttp/cli/internal/output"
	"github.com/benchttp/cli/internal/render"
	"github.com/benchttp/cli/internal/signals"
)

// cmdRun handles subcommand "benchttp run [options]".
type cmdRun struct {
	flagset *flag.FlagSet

	// configFile is the parsed value for flag -configFile
	configFile string

	// config is the runner config resulting from parsing CLI flags.
	config runner.Config
}

// init initializes cmdRun with default values.
func (cmd *cmdRun) init() {
	cmd.config = runner.DefaultConfig()
	cmd.configFile = configfile.Find([]string{
		"./.benchttp.yml",
		"./.benchttp.yaml",
		"./.benchttp.json",
	})
}

// execute runs the benchttp runner: it parses CLI flags, loads config
// from config file and parsed flags, then runs the benchmark and outputs
// it according to the config.
func (cmd *cmdRun) execute(args []string) error {
	cmd.init()

	// Generate merged config (default < config file < CLI flags)
	cfg, err := cmd.makeConfig(args)
	if err != nil {
		return err
	}

	report, err := runBenchmark(cfg)
	if err != nil {
		return err
	}

	return renderReport(os.Stdout, report, cfg.Output.Silent)
}

// parseArgs parses input args as config fields and returns
// a slice of fields that were set by the user.
func (cmd *cmdRun) parseArgs(args []string) []string {
	// first arg is subcommand "run"
	// skip parsing if no flags are provided
	if len(args) <= 1 {
		return []string{}
	}

	// config file path
	cmd.flagset.StringVar(&cmd.configFile,
		"configFile",
		cmd.configFile,
		"Config file path",
	)

	// attach config options flags to the flagset
	// and bind their value to the config struct
	configflag.Bind(cmd.flagset, &cmd.config)

	cmd.flagset.Parse(args[1:]) //nolint:errcheck // never occurs due to flag.ExitOnError

	return configflag.Which(cmd.flagset)
}

// makeConfig returns a runner.ConfigGlobal initialized with config file
// options if found, overridden with CLI options listed in fields
// slice param.
func (cmd *cmdRun) makeConfig(args []string) (cfg runner.Config, err error) {
	// Set CLI config from flags and retrieve fields that were set
	fields := cmd.parseArgs(args)

	// configFile not set and default ones not found:
	// skip the merge and return the cli config
	if cmd.configFile == "" {
		return cmd.config, cmd.config.Validate()
	}

	fileConfig, err := configfile.Parse(cmd.configFile)
	if err != nil && !errors.Is(err, configfile.ErrFileNotFound) {
		// config file is not mandatory: discard ErrFileNotFound.
		// other errors are critical
		return
	}

	mergedConfig := fileConfig.Override(cmd.config, fields...)

	return mergedConfig, mergedConfig.Validate()
}

func onRecordingProgress(silent bool) func(runner.RecordingProgress) {
	if silent {
		return func(runner.RecordingProgress) {}
	}

	// hack: write a blank line as render.Progress always
	// erases the previous line
	fmt.Println()

	return func(progress runner.RecordingProgress) {
		render.Progress(os.Stdout, progress) //nolint: errcheck
	}
}

func runBenchmark(cfg runner.Config) (*runner.Report, error) {
	// Prepare graceful shutdown in case of os.Interrupt (Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	go signals.ListenOSInterrupt(cancel)

	// Run the benchmark
	report, err := runner.
		New(onRecordingProgress(cfg.Output.Silent)).
		Run(ctx, cfg)
	if err != nil {
		return report, err
	}

	return report, nil
}

func renderReport(w io.Writer, report *runner.Report, silent bool) error {
	cw := output.ConditionalWriter{Writer: w}.If(!silent)

	if _, err := render.ReportSummary(cw, report); err != nil {
		return err
	}

	if _, err := render.TestSuite(
		cw.Or(!report.Tests.Pass), report.Tests,
	); err != nil {
		return err
	}

	if !report.Tests.Pass {
		return errors.New("test suite failed")
	}

	return nil
}
