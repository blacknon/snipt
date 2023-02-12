// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"time"

	"github.com/blacknon/snipt/client"
	"github.com/urfave/cli/v2"
)

// CmdCreate
var CmdCreate = cli.Command{
	Name:      "create",
	Usage:     "create remote snippet. default by github creates a secret gist, gitlab snippet creates a private snippet.",
	Action:    cmdActionCreate,
	ArgsUsage: "FILE...",
	Flags: []cli.Flag{
		// -v
		CommonFlagSelecterVisibility,

		// -t
		CommonFlagSetTitle,

		// -p
		&cli.BoolFlag{
			Name:    "project_snippet",
			Aliases: []string{"p"},
			Usage:   "output to a list so that it can also support the creation of Gitlab's Project Snippet.",
		},

		// -A
		// &cli.BoolFlag{
		// 	Name:    "ask",
		// 	Aliases: []string{"T"},
		// 	Usage:   "",
		// },
	},
}

func cmdActionCreate(c *cli.Context) (err error) {
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

	// Select platform to create snippet
	platformList, err := cl.PlatformList(c.Bool("project_snippet"))
	if err != nil {
		return err
	}

	var filterText string
	for _, p := range platformList {
		t := fmt.Sprintln(p)
		filterText += t
	}

	// Run filter command
	text, err := filter(conf.General.SelectCmd, []string{}, filterText)
	if err != nil {
		return
	}

	var rawURLs []string
	for _, t := range text {
		// set title
		title := c.String("title")
		if title == "" {
			timestamp := time.Now().Format("2006/01/02 15:04:05")
			title = fmt.Sprintf("Snippet at %s", timestamp)
		}

		// Create Snippet
		snippetData := client.SnippetData{
			Title: title,
			Files: snippetFileDataList,
		}

		// set visibility
		if c.Bool("visibility") {

			// Get platorm visibility list
			vl := cl.VisibilityListFromPlatform(t)

			// select visibility
			visibility, eErr := getSelectVisibility(vl)

			if eErr != nil {
				return err
			}

			snippetData.Visibility = visibility
		}

		rawURL, eErr := cl.Create(t, snippetData)
		if err != nil {
			return eErr
		}

		rawURLs = append(rawURLs, rawURL...)

	}

	for _, url := range rawURLs {
		fmt.Printf("Snippet created: %s\n", url)
	}

	return
}
