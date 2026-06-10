package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

// runInit writes a starter config with example mcpServers entries (disabled)
// to the global config path, or to -path if given. Existing files are left
// untouched.
func runInit(args []string, logger *slog.Logger) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var pathOpt string
	fs.StringVar(&pathOpt, "path", "", "config file to create; defaults to the global user config")

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

	path := pathOpt
	if path == "" {
		var err error
		path, err = config.DefaultGlobalPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	created, err := config.WriteStarter(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if created {
		logger.Info("created starter config", "path", path)
		fmt.Fprintf(os.Stdout, "created %s\n", path)
	} else {
		logger.Info("config already exists", "path", path)
		fmt.Fprintf(os.Stdout, "exists %s\n", path)
	}

	return 0
}
