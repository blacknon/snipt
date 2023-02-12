// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blacknon/snipt/client"
	"github.com/urfave/cli/v2"
)

// CmdGet
var CmdGet = cli.Command{
	Name:   "get",
	Usage:  "get remote snippet data.",
	Action: cmdActionGet,
	Flags: []cli.Flag{
		// -o PATH
		CommonFlagOutput,

		// -f
		CommonFlagSnippetFile,

		// -s
		CommonFlagViewSecret,

		// -r
		&cli.BoolFlag{
			Name:    "read",
			Aliases: []string{"r"},
			Usage:   "printout to stdout from snippet.",
		},
	},
}

// actionList is the function that defines the processing of the lists subcommand.
func cmdActionGet(c *cli.Context) (err error) {
	// Get **config data** and **client.Client**
	cf := c.String("config")
	conf, cl, err := clinetInit(cf)
	if err != nil {
		return
	}

	// Get List
	list := cl.List(c.Bool("file"), c.Bool("secret"))

	// Create list
	var filterText string
	for _, l := range list {
		t := fmt.Sprintln(l.URL, l.Platform, l.Description)
		filterText += t
	}

	// Run filter command
	text, err := filter(conf.General.SelectCmd, []string{}, filterText)

	// get snippet files
	files := []client.SnippetFileData{}
	for _, t := range text {
		// generate url as search key value.
		splitText := strings.Split(t, " ")
		url := splitText[0]

		// Get SnippetData
		snippet, err := cl.Get(url)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		// append file
		if c.Bool("file") {
			for _, f := range snippet.Files {
				if url == f.Filter {
					files = append(files, f)
				}
			}
		} else {
			files = append(files, snippet.Files...)
		}

	}

	// check multiple file
	isMultiple := false
	if len(files) > 1 {
		isMultiple = true
	}

	// write data
	for _, file := range files {
		err = outputGetData(c.Bool("read"), isMultiple, c.String("output"), file)
		if err != nil {
			return
		}
	}

	return
}

// outputGetData
func outputGetData(isRead, isMultiple bool, dir string, file client.SnippetFileData) (err error) {
	p := file.Path
	if dir != "" {
		p = filepath.Join(dir, p)
	}

	// get path
	path := getFullPath(p)

	// set writer in os.Stdout
	w := os.Stdout
	if !isRead {
		// file exist check
		if isExist(path) {
			if !askYesNo(fmt.Sprintf("Overwrite %s ?", path)) {
				return
			}
		}

		// open file and set writer
		w, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return
		}
		defer w.Close()
	}

	err = write(w, file.Contents)

	return
}
