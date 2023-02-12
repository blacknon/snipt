// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli/v2"
)

// App
var App = &cli.App{
	Name:      "snipt",
	Usage:     "multiple remote platform snippet manager.",
	Version:   "0.1.0",
	ErrWriter: ioutil.Discard,

	// Flags
	Flags: commonFlags,

	// Commands
	Commands: []*cli.Command{
		// list subcommand
		&CmdList,

		// get subcommand
		&CmdGet,

		// create subcommand
		&CmdCreate,

		// update subcommand
		&CmdUpdate,

		// edit subcommand
		&CmdEdit,

		// delete subcommand
		&CmdDelete,

		// add subcommand

		// comment subcommand
	},

	// Output usages and error messages
	OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
		cli.ShowAppHelp(c)
		return err
	},
}

// CommonFlags
var commonFlags = []cli.Flag{
	// config option
	&cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "load configuration from `FILE`",
	},
}

// CommonFlagOutput ... -o, --output
var CommonFlagOutput = &cli.StringFlag{
	Name:    "output",
	Aliases: []string{"o"},
	Usage:   "output snippet to `PATH`",
}

// CommonFlagSetTitle ... -t, --title
var CommonFlagSetTitle = &cli.StringFlag{
	Name:    "title",
	Aliases: []string{"t"},
	Usage:   "specify remote snippet title.",
}

// CommonFlagSnippetFile ... -f, --file
var CommonFlagSnippetFile = &cli.BoolFlag{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "output snippet by file",
}

// CommonFlagViewSecret .. -s, --secret
var CommonFlagViewSecret = &cli.BoolFlag{
	Name:    "secret",
	Aliases: []string{"s"},
	Usage:   "printout",
}

// CommonFlagSelecterVisibility ... -v, --visibility
var CommonFlagSelecterVisibility = &cli.BoolFlag{
	Name:    "visibility",
	Aliases: []string{"v"},
	Usage:   "specify visibility according to each `github gist`/`gitlab snippet`.",
}

// Execute
func Execute() {
	// execute command
	if err := App.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}
