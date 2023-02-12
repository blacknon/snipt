// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	// "github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

// CmdList
var CmdList = cli.Command{
	Name:   "list",
	Usage:  "list all snippet.",
	Action: cmdActionList,
	Flags: []cli.Flag{
		// -T
		// &cli.BoolFlag{
		// 	Name:    "table",
		// 	Aliases: []string{"T"},
		// 	Usage:   "output in markdown table format",
		// },

		// -f
		CommonFlagSnippetFile,

		// -s
		CommonFlagViewSecret,
	},
}

// actionList is the function that defines the processing of the lists subcommand.
func cmdActionList(c *cli.Context) (err error) {
	// Get **config data** and **client.Client**
	cf := c.String("config")
	_, cl, err := clinetInit(cf)
	if err != nil {
		return
	}

	// Get List
	list := cl.List(c.Bool("file"), c.Bool("secret"))

	// Output list
	for _, l := range list {
		u := l.URL
		if c.Bool("file") {
			u = l.RawURL
		}

		t := fmt.Sprintln(u, l.Visibility+": "+l.Description)
		fmt.Print(t)
	}

	return
}
