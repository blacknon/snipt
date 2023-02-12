// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import "github.com/urfave/cli/v2"

// CmdAdd
var CmdAdd = cli.Command{
	Name:   "add",
	Usage:  "add snippet file to remote snippet.",
	Action: cmdActionAdd,
}

func cmdActionAdd(c *cli.Context) (err error) {

	return
}
