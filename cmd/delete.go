// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

// CmdDelete
var CmdDelete = cli.Command{
	Name:   "delete",
	Usage:  "delete remote snippet data.",
	Action: cmdActionDelete,
	Flags: []cli.Flag{
		// -s
		CommonFlagViewSecret,
	},
}

func cmdActionDelete(c *cli.Context) (err error) {
	// Get **config data** and **client.Client**
	cf := c.String("config")
	conf, cl, err := clinetInit(cf)
	if err != nil {
		return
	}

	// Get List
	list := cl.List(false, c.Bool("secret"))

	// Create list
	var filterText string
	for _, l := range list {
		t := fmt.Sprintln(l.URL, l.Platform, l.Description)
		filterText += t
	}

	// Run filter command
	text, err := filter(conf.General.SelectCmd, []string{}, filterText)

	for _, t := range text {
		splitText := strings.Split(t, " ")
		url := splitText[0]

		err := cl.Delete(url)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		fmt.Fprintf(os.Stderr, "Snippet deleted: %s\n", url)
	}

	return
}
