package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/benchttp/engine/configparse"
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

	// silent is the parsed value for flag -silent
	silent bool

	// configRepr is the runner config resulting from config flag values
	configRepr configparse.Representation
}

// execute runs the benchttp runner: it parses CLI flags, loads config
// from config file and parsed flags, then runs the benchmark and outputs
// it according to the config.
func (cmd *cmdRun) execute(args []string) error {
	if err := cmd.parseArgs(args); err != nil {
		return err
	}

	config, err := buildConfig(cmd.configFile, cmd.configRepr)
	if err != nil {
		return err
	}

	report, err := runBenchmark(config, cmd.silent)
	if err != nil {
		return err
	}

	return renderReport(os.Stdout, report, cmd.silent)
}

func (cmd *cmdRun) parseArgs(args []string) error {
	cmd.flagset.StringVar(&cmd.configFile, "configFile", configfile.Find(), "Config file path")
	cmd.flagset.BoolVar(&cmd.silent, "silent", false, "Silent mode")
	configflag.Bind(cmd.flagset, &cmd.configRepr)
	return cmd.flagset.Parse(args)
}

func buildConfig(
	filePath string,
	cliConfigRepr configparse.Representation,
) (runner.Config, error) {
	// start with default config as base
	config := runner.DefaultConfig()

	// override with config file values
	err := configfile.Parse(filePath, &config)
	if err != nil && !errors.Is(err, configfile.ErrFileNotFound) {
		// config file is not mandatory: discard ErrFileNotFound.
		// other errors are critical
		return config, err
	}

	// override with CLI flags values
	if err := cliConfigRepr.ParseInto(&config); err != nil {
		return config, err
	}

	return config, nil
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

func runBenchmark(cfg runner.Config, silent bool) (*runner.Report, error) {
	// Prepare graceful shutdown in case of os.Interrupt (Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	go signals.ListenOSInterrupt(cancel)

	// Run the benchmark
	report, err := runner.
		New(onRecordingProgress(silent)).
		Run(ctx, cfg)
	if err != nil {
		return report, err
	}

	return report, nil
}

func renderReport(w io.Writer, report *runner.Report, silent bool) error {
	writeIfNotSilent := output.ConditionalWriter{Writer: w}.If(!silent)

	if _, err := render.ReportSummary(writeIfNotSilent, report); err != nil {
		return err
	}

	if _, err := render.TestSuite(
		writeIfNotSilent.ElseIf(!report.Tests.Pass),
		report.Tests,
	); err != nil {
		return err
	}

	if !report.Tests.Pass {
		return errors.New("test suite failed")
	}

	return nil
}
