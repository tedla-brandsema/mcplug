package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

type initOptions struct {
	pathOpt string
}

func runInit(args []string, logger *slog.Logger) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var opts initOptions
	fs.StringVar(&opts.pathOpt, "path", "", "config file to create; defaults to the global user config")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		return 2
	}

	_, configPath, err := config.LoadOrCreate(opts.pathOpt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	logger.Info("config ready", "path", configPath)
	fmt.Fprintf(os.Stdout, "config at %s\n", configPath)

	return 0
}
