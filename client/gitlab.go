// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/xanzy/go-gitlab"
)

type GitlabClient struct {
	ctx          context.Context
	client       *gitlab.Client
	Url          string
	User         string
	PlatformName string
	FilterKey    string
	Project      *gitlab.Project

	// proxy
	proxy     string
	proxyUser string
	proxyPass string
}

var (
	GitlabIsPrivate  = Visibility{code: "private", num: 0}
	GitlabIsPublic   = Visibility{code: "public", num: 1}
	GitlabIsInternal = Visibility{code: "internal", num: 2}
)

// Init
func (g *GitlabClient) Init(u, token string) (err error) {
	// create ctx
	g.ctx = context.Background()

	transport := &http.Transport{}
	if g.proxy != "" {
		proxyUrl, err := url.Parse(g.proxy)
		if err != nil {
			return err
		}

		hdr := make(http.Header)
		hdr.Add("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(g.proxyUser+":"+g.proxyPass)))

		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	h := &http.Client{Transport: transport}

	// Create Gitlab Client
	g.client, err = gitlab.NewClient(token, gitlab.WithBaseURL(u), gitlab.WithHTTPClient(h))
	if err != nil {
		return
	}

	// set url
	g.Url = u

	// set username
	user, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return
	}
	g.User = user.Name

	// Generate PlatformName
	g.PlatformName = g.client.BaseURL().Host + ":" + g.User

	return
}

// List
func (g *GitlabClient) List(isFile, isSecret bool) (snippetList SnippetList, err error) {
	// set ListProjectsOptions pageSize
	const pageSize = 50

	opt := gitlab.ListSnippetsOptions(
		gitlab.ListOptions{
			Page:    0,
			PerPage: pageSize,
		},
	)

	for {
		// get snippetList
		snippetDataList, resp, ferr := g.client.Snippets.ListSnippets(&opt)

		// check error
		if ferr != nil {
			return snippetList, ferr
		}

		for _, snippet := range snippetDataList {
			if !isSecret {
				//
				v := getGitlabVisibilityFromString(snippet.Visibility)
				switch v {
				case GitlabIsPrivate, GitlabIsInternal:
					continue
				}
			}

			// get Description
			title := replaceNewline(snippet.Title, "\\n")

			data := SnippetListData{
				Client:     g,
				Platform:   g.PlatformName,
				Id:         strconv.Itoa(snippet.ID),
				Title:      title,
				URL:        snippet.WebURL,
				Visibility: snippet.Visibility,
			}

			// check file flag
			if isFile {
				if len(snippet.Files) > 1 {
					for _, f := range snippet.Files {
						fd := data
						fd.URL, _ = url.JoinPath(fd.URL, f.Path)
						fd.RawURL = f.RawURL

						snippetList = append(snippetList, &fd)
					}
				} else {
					data.URL, _ = url.JoinPath(data.URL, snippet.FileName)
					data.RawURL = snippet.RawURL
					snippetList = append(snippetList, &data)
				}
			} else {
				snippetList = append(snippetList, &data)
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return
}

// Get
func (g *GitlabClient) Get(id string) (snippet SnippetData, err error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		return
	}

	sn, _, err := g.client.Snippets.GetSnippet(intId)
	if err != nil {
		return
	}

	snippetFiles := sn.Files
	files := []SnippetFileData{}
	if len(snippetFiles) > 1 {
		for _, f := range snippetFiles {
			ref := "main"
			contentByte, _, err := g.client.Snippets.SnippetFileContent(intId, ref, f.Path)
			if err != nil {
				fmt.Println(err)
				continue
			}

			filterVal := sn.WebURL + "/" + f.Path

			fd := SnippetFileData{
				Filter:   filterVal,
				RawURL:   f.RawURL,
				Path:     f.Path,
				Contents: contentByte,
			}

			files = append(files, fd)
		}
	} else {
		contentByte, _, _ := g.client.Snippets.SnippetContent(intId)
		filterVal := sn.WebURL + "/" + sn.FileName
		fd := SnippetFileData{
			Filter:   filterVal,
			RawURL:   sn.RawURL,
			Path:     sn.FileName,
			Contents: contentByte,
		}

		files = append(files, fd)
	}

	var visibility Visibility
	switch sn.Visibility {
	case "private":
		visibility = GitlabIsPrivate
	case "public":
		visibility = GitlabIsPublic
	case "internal":
		visibility = GitlabIsInternal
	default:
	}

	snippet = SnippetData{
		Title:       sn.Title,
		Description: sn.Description,
		URL:         sn.WebURL,
		Visibility:  visibility,
		Files:       files,
	}

	return
}

// Create
func (g *GitlabClient) Create(data SnippetData) (snippet SnippetClient, err error) {
	// set default visiblity
	if data.Visibility == (Visibility{}) {
		data.Visibility = GitlabIsPrivate
	}

	// create files
	files, fileName, contents := createGitlabCreateSnippetFiles(data.Files)

	// set visibility
	visibility := getGitlabVisibility(data.Visibility)

	if g.Project == nil {
		// create opt
		opt := &gitlab.CreateSnippetOptions{}
		opt.Title = gitlab.String(data.Title)
		opt.Description = gitlab.String(data.Description)
		opt.Visibility = gitlab.Visibility(visibility)

		if len(files) > 1 {
			opt.Files = &files
		} else {
			opt.FileName = &fileName
			opt.Content = &contents
		}

		snippet, _, err = g.client.Snippets.CreateSnippet(opt)

	} else {
		// create opt
		opt := &gitlab.CreateProjectSnippetOptions{}
		opt.Title = gitlab.String(data.Title)
		opt.Description = gitlab.String(data.Description)
		opt.Visibility = gitlab.Visibility(visibility)

		if len(files) > 1 {
			opt.Files = &files
		} else {
			opt.FileName = &fileName
			opt.Content = &contents
		}

		snippet, _, err = g.client.ProjectSnippets.CreateSnippet(g.Project.ID, opt)
	}

	return
}

// Update
func (g *GitlabClient) Update(id string, data SnippetData) (snippet SnippetClient, err error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		return
	}

	// create files
	files, fileName, contents := createGitlabUpdateSnippetFiles(data.Files)

	// set visibility
	visibility := getGitlabVisibility(data.Visibility)

	// TODO: filesがstructにないっぽいので、PR出す
	if g.Project == nil {
		// create createOpt
		opt := &gitlab.UpdateSnippetOptions{}
		opt.Title = gitlab.String(data.Title)
		opt.Description = gitlab.String(data.Description)
		opt.Visibility = &visibility

		if len(files) > 1 {
			opt.Files = &files
		} else {
			opt.FileName = &fileName
			opt.Content = &contents
		}

		snippet, _, err = g.client.Snippets.UpdateSnippet(intId, opt)
	} else {
		opt := &gitlab.UpdateProjectSnippetOptions{}
		opt.Title = gitlab.String(data.Title)
		opt.Description = gitlab.String(data.Description)
		opt.Visibility = &visibility

		if len(files) > 1 {
			opt.Files = &files
		} else {
			opt.FileName = &fileName
			opt.Content = &contents
		}

		snippet, _, err = g.client.ProjectSnippets.UpdateSnippet(g.Project.ID, intId, opt)
	}

	return
}

// Delete
func (g *GitlabClient) Delete(strId string) (err error) {
	id, err := strconv.Atoi(strId)
	if err != nil {
		return
	}

	_, err = g.client.Snippets.DeleteSnippet(id)
	return
}

// GetPlatformName
func (g *GitlabClient) GetPlatformName() string {
	return g.PlatformName
}

// GetProjectList
func (g *GitlabClient) GetProjectList() (projectList []*gitlab.Project, err error) {
	// set ListProjectsOptions pageSize
	const pageSize = 50

	opt := gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    0,
			PerPage: pageSize,
		},
		Archived:   boolAddr(false),
		Membership: boolAddr(true),
	}

	for {
		projects, resp, err := g.client.Projects.ListProjects(&opt)
		if err != nil {
			return []*gitlab.Project{}, err
		}

		for _, p := range projects {
			if p.SnippetsEnabled {
				projectList = append(projectList, p)
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.ListOptions.Page = resp.NextPage
	}

	return
}

// GetFilterKey
func (g *GitlabClient) GetFilterKey() string {
	return g.FilterKey
}

// SetFilterKey
func (g *GitlabClient) SetFilterKey(key string) {
	g.FilterKey = key
}

// VisibilityList
func (g *GitlabClient) VisibilityList() (visibilityList []Visibility) {
	visibilityList = []Visibility{
		GitlabIsPrivate,
		GitlabIsPublic,
		GitlabIsInternal,
	}

	return
}

// createGistFile
func createGitlabCreateSnippetFiles(data []SnippetFileData) (files []*gitlab.CreateSnippetFileOptions, fileName, contents string) {
	// set data to files
	i := 0
	for _, d := range data {
		// []byte to string
		c := string(d.Contents)
		filepath := d.Path

		// create gitlab.SnippetFile
		f := gitlab.CreateSnippetFileOptions{
			FilePath: &filepath,
			Content:  &c,
		}

		files = append(files, &f)

		if i == 0 {
			fileName = d.Path
			contents = c
		}
		i++
	}

	return
}

// createGistFile
func createGitlabUpdateSnippetFiles(data []SnippetFileData) (files []*gitlab.UpdateSnippetFileOptions, fileName, contents string) {
	// set data to files
	i := 0
	for _, d := range data {
		// []byte to string
		c := string(d.Contents)
		filepath := d.Path

		f := gitlab.UpdateSnippetFileOptions{
			Action:   gitlab.String("update"),
			FilePath: &filepath,
			Content:  &c,
		}

		files = append(files, &f)

		if i == 0 {
			fileName = d.Path
			contents = c
		}

		i++
	}

	return
}

// getGitlabVisibility
func getGitlabVisibility(v Visibility) (visibility gitlab.VisibilityValue) {
	switch v {
	case GitlabIsPrivate:
		visibility = gitlab.PrivateVisibility
	case GitlabIsPublic:
		visibility = gitlab.PublicVisibility
	case GitlabIsInternal:
		visibility = gitlab.InternalVisibility
	}

	return
}

// getGitlabVisibilityFromString
func getGitlabVisibilityFromString(s string) (v Visibility) {
	switch s {
	case GitlabIsPrivate.GetCode():
		v = GitlabIsPrivate
	case GitlabIsPublic.GetCode():
		v = GitlabIsPublic
	case GitlabIsInternal.GetCode():
		v = GitlabIsInternal
	}

	return
}
