// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/blacknon/snipt/client"
)

func askYesNo(s string) (result bool) {
	message := fmt.Sprintf("%s [y/n]: ", s)

	qs := []*survey.Question{
		{
			Name: "yesno",
			Prompt: &survey.Select{
				Message: message,
				Options: []string{"yes", "no"},
			},
		},
	}

	answers := struct {
		YesNo string `survey:"yesno"`
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		return false
	}

	result = false
	if answers.YesNo == "yes" {
		result = true
	}

	return
}

func askChoose(s string, list []string) (result string, err error) {
	message := fmt.Sprintf("%s: ", s)

	qs := []*survey.Question{
		{
			Name: "choose",
			Prompt: &survey.Select{
				Message: message,
				Options: list,
			},
		},
	}

	answers := struct {
		Choose string `survey:"choose"`
	}{}

	err = survey.Ask(qs, &answers)
	result = answers.Choose

	return
}

func askInput(s string) (result string, err error) {
	message := fmt.Sprintf("%s: ", s)

	qs := []*survey.Question{
		{
			Name: "input",
			Prompt: &survey.Input{
				Message: message,
			},
			Validate:  survey.Required,
			Transform: survey.Title,
		},
	}

	answers := struct {
		Input string `survey:"input"`
	}{}

	err = survey.Ask(qs, &answers)
	result = answers.Input

	return
}

func isExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func run(command string, r io.Reader, w io.Writer) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = w
	cmd.Stdin = r
	return cmd.Run()
}

func filter(filter string, filterOptions []string, filterText string) (filteredData []string, err error) {
	//
	var buf bytes.Buffer
	selectCmd := fmt.Sprintf("%s %s", filter, strings.Join(filterOptions, " "))

	//
	err = run(selectCmd, strings.NewReader(filterText), &buf)
	if err != nil {
		return
	}

	// get filteredData.
	filteredData = strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")

	return
}

// write
func write(w *os.File, data []byte) (err error) {
	// write file
	_, err = w.Write(data)
	if err != nil {
		return
	}

	// close file
	err = w.Close()

	return
}

// edit
func edit(editor string, editorOptions []string, filename string, data []byte) (tmpfilePath string, err error) {
	// create tmpfile name
	rand.Seed(time.Now().Unix())
	tmpfileName := "tmp_" + time.Now().Format("20060102150405")
	tmpfileName = tmpfileName + "_" + filename

	// create tmpfile
	tmpfile, err := os.CreateTemp(os.TempDir(), tmpfileName)
	tmpfilePath = tmpfile.Name()

	// create tmpfile
	err = write(tmpfile, data)
	if err != nil {
		return
	}

	// create editor command
	editorCmd := fmt.Sprintf("%s %s %s", editor, strings.Join(editorOptions, " "), tmpfilePath)

	// run editor command
	err = run(editorCmd, os.Stdin, os.Stdout)
	if err != nil {
		return
	}

	return
}

func read(path string) (data []byte, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	return bs, nil
}

// GetFullPath returns a fullpath of path.
// Expands `~` to user directory ($HOME environment variable).
func getFullPath(path string) (fullPath string) {
	if path == "" {
		fullPath = path
	} else {
		usr, _ := user.Current()
		fullPath = strings.Replace(path, "~", usr.HomeDir, 1)
		fullPath, _ = filepath.Abs(fullPath)
	}
	return fullPath
}

// createSnippetData
func createSnippetData(pathList []string) (snippetDataList []client.SnippetFileData, err error) {
	for _, path := range pathList {
		content, err := read(path)
		if err != nil {
			return snippetDataList, err
		}

		data := client.SnippetFileData{
			Path:     filepath.Base(path),
			Contents: content,
		}

		snippetDataList = append(snippetDataList, data)
	}

	return snippetDataList, err
}

func getSelectVisibility(visibilityList []client.Visibility) (visibility client.Visibility, err error) {
	msg := "select remote snippet visibility"

	selectList := []string{}
	for _, v := range visibilityList {
		selectList = append(selectList, v.GetCode())
	}

	r, err := askChoose(msg, selectList)
	if err != nil {
		return
	}

	for _, v := range visibilityList {
		if v.GetCode() == r {
			visibility = v
			break
		}
	}

	return visibility, err
}
