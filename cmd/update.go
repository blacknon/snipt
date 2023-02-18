// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/blacknon/snipt/client"
	"github.com/urfave/cli/v2"
)

// CmdUpdate
var CmdUpdate = cli.Command{
	Name:   "update",
	Usage:  "update remote snippet data.",
	Action: cmdActionUpdate,
	Flags: []cli.Flag{
		// -f
		CommonFlagSnippetFile,

		// -v
		CommonFlagSelecterVisibility,

		// -t
		CommonFlagSetTitle,

		// -s
		CommonFlagViewSecret,
	},
}

// TODO: renameの処理を実装する(オプション？)

func cmdActionUpdate(c *cli.Context) (err error) {
	// check args count
	if c.NArg() == 0 {
		err = fmt.Errorf("no arguments")
		c.App.OnUsageError(c, err, true)
		return
	}

	// get args
	args := c.Args().Slice()
	pathList, err := getPathList(args)
	if err != nil {
		return
	}

	// generate SnippetData from pathList
	snippetFileDataList, err := createSnippetData(pathList)
	if err != nil {
		return
	}

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
		t := fmt.Sprintln(l.URL, l.Platform, l.Title)
		filterText += t
	}

	// Run filter command
	text, err := filter(conf.General.SelectCmd, []string{}, filterText)
	if err != nil {
		return
	}

	// set title
	title := c.String("title")

	urlList := []string{}
	for _, t := range text {
		// generate url as search key value.
		splitText := strings.Split(t, " ")
		url := splitText[0]

		// Get SnippetData
		snippetData, err := cl.Get(url)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		fmt.Println("- 1. title: ", snippetData.Title)

		// Get and Set title
		if title != "" {
			snippetData.Title = title
		}

		fmt.Println("- 2. title: ", snippetData.Title)

		// update visibility
		if c.Bool("visibility") {
			// Get platorm visibility list
			vl, eErr := cl.VisibilityListFromURL(url)
			if eErr != nil {
				return err
			}

			// select visibility
			visibility, eErr := getSelectVisibility(vl)
			if eErr != nil {
				return err
			}

			snippetData.Visibility = visibility
		}

		// append file
		files := []client.SnippetFileData{}
		for _, f := range snippetFileDataList {
			for _, sf := range snippetData.Files {
				if f.Path == sf.Path {
					files = append(files, f)
				} else {
					files = append(files, sf)
				}
			}
		}
		snippetData.Files = files

		fmt.Println("- 3. title: ", snippetData.Title)

		// update
		rawURLs, err := cl.Update(url, snippetData)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		urlList = append(urlList, rawURLs...)
	}

	for _, u := range urlList {
		fmt.Printf("Snippet Update: %s\n", u)
	}

	return
}
