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

// CmdEdit
var CmdEdit = cli.Command{
	Name:   "edit",
	Usage:  "edit remote snippet file. use the command specified in `editor` in config.toml for editing.",
	Action: cmdActionEdit,
	Flags: []cli.Flag{
		// -v
		CommonFlagSelecterVisibility,

		// -t
		CommonFlagSetTitle,

		// -s
		CommonFlagViewSecret,
	},
}

// TODO: renameの処理を実装する(オプション？)

func cmdActionEdit(c *cli.Context) (err error) {
	// Get **config data** and **client.Client**
	cf := c.String("config")
	conf, cl, err := clinetInit(cf)
	if err != nil {
		return
	}

	// Get List
	list := cl.List(true, c.Bool("secret"))

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

	// TODO: 1ライン以上選択されている場合はエラーにする？
	if len(text) == 0 {
		return
	}

	urlList := []string{}
	for _, t := range text {
		// create url
		splitText := strings.Split(t, " ")
		url := splitText[0]

		snippetData, eErr := cl.Get(url)
		if eErr != nil {
			fmt.Fprintln(os.Stderr, eErr)
			return err
		}

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

		// edit
		editedFiles, eErr := editFiles(url, conf.General.Editor, []string{}, snippetData.Files)
		if eErr != nil {
			return
		}
		snippetData.Files = editedFiles

		// edit data update
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

func editFiles(url, editor string, editorOptions []string, files []client.SnippetFileData) (editedFiles []client.SnippetFileData, err error) {
	for _, f := range files {
		if url == f.Filter {
			editedPathList := []string{}
			tmpfile, eerr := edit(editor, editorOptions, f.Path, f.Contents)
			if eerr != nil {
				return editedFiles, eerr
			}
			defer os.Remove(tmpfile)

			// create SnippetData list...
			editedPathList = append(editedPathList, tmpfile)
			tmpEditedFiles, eerr := createSnippetData(editedPathList)
			if eerr != nil {
				return editedFiles, eerr
			}

			// update path
			tmpEditedFiles[0].Path = f.Path

			editedFiles = append(editedFiles, tmpEditedFiles...)
		} else {
			editedFiles = append(editedFiles, f)
		}
	}

	return
}
