// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package client

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GistClient struct {
	ctx          context.Context
	client       *github.Client
	User         string
	FilterKey    string
	PlatformName string
}

var (
	GistIsSecret = Visibility{code: "secret", num: 0}
	GistIsPublic = Visibility{code: "public", num: 1}
)

// Init
func (g *GistClient) Init(token string) (err error) {
	// create ctx
	g.ctx = context.Background()

	// Create oAuth2 Client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(g.ctx, ts)

	// Create Client
	g.client = github.NewClient(tc)

	// Get login user
	user, _, err := g.client.Users.Get(g.ctx, "")
	if err != nil {
		return
	}
	g.User = *user.Login

	// Generate PlatformName
	// host := g.client.BaseURL.Host
	host := "gist.github.com"
	g.PlatformName = fmt.Sprintf("%s:%s", host, g.User)

	return
}

// List
func (g *GistClient) List(isFile, isSecret bool) (snippetList SnippetList, err error) {
	// create gist list options
	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// get gistList
	gistDataList, _, err := g.client.Gists.List(g.ctx, "", opt)

	// create []SnipetListData
	for _, gist := range gistDataList {
		if !isSecret && !gist.GetPublic() {
			continue
		}

		// get Description
		description := replaceNewline(gist.GetDescription(), "\\n")

		// get visibility
		visibility := "secret"
		if gist.GetPublic() {
			visibility = "public"
		}

		data := SnippetListData{
			Client:     g,
			Platform:   g.PlatformName,
			Id:         gist.GetID(),
			Title:      description,
			URL:        gist.GetHTMLURL(),
			Visibility: visibility,
		}

		if isFile {
			for _, f := range gist.Files {
				fd := data
				fd.URL, _ = url.JoinPath(fd.URL, f.GetFilename())
				fd.RawURL = f.GetRawURL()
				snippetList = append(snippetList, &fd)
			}
		} else {
			snippetList = append(snippetList, &data)
		}
	}

	sort.Slice(snippetList, func(i, j int) bool {
		// i番目とj番目の要素のAgeを比較
		return snippetList[i].URL < snippetList[j].URL
	})

	return snippetList, err
}

// Get
func (g *GistClient) Get(id string) (data SnippetData, err error) {
	gist, _, err := g.client.Gists.Get(g.ctx, id)

	files := []SnippetFileData{}
	for _, file := range gist.Files {

		fd := SnippetFileData{
			Filter:   gist.GetHTMLURL() + "/" + file.GetFilename(),
			RawURL:   file.GetRawURL(),
			Path:     file.GetFilename(),
			Contents: []byte(file.GetContent()),
		}

		files = append(files, fd)
	}

	visibility := GistIsSecret
	if gist.GetPublic() {
		visibility = GistIsPublic
	}

	data = SnippetData{
		Title:      gist.GetDescription(),
		URL:        gist.GetHTMLURL(),
		Visibility: visibility,
		Files:      files,
	}

	return
}

// Create
func (g *GistClient) Create(data SnippetData) (gist SnippetClient, err error) {
	// set default visiblity
	if data.Visibility == (Visibility{}) {
		data.Visibility = GistIsSecret
	}

	// set isPublic
	isPublic := false
	if data.Visibility == GistIsPublic {
		isPublic = true
	}

	// create files
	files := createGithubGistFiles(data.Files)

	// create gist
	gist, _, err = g.client.Gists.Create(
		g.ctx,
		&github.Gist{
			Description: &data.Title,
			Files:       files,
			Public:      &isPublic,
		})

	return gist, err
}

// Update
func (g *GistClient) Update(id string, data SnippetData) (gist SnippetClient, err error) {
	// set visiblity
	isPublic := false
	if data.Visibility == GistIsPublic {
		isPublic = true
	}

	// create files
	files := createGithubGistFiles(data.Files)

	// update gist
	gist, _, err = g.client.Gists.Edit(
		g.ctx,
		id,
		&github.Gist{
			Description: &data.Title,
			Files:       files,
			Public:      &isPublic,
		})

	return gist, err
}

// Delete
func (g *GistClient) Delete(id string) (err error) {
	_, err = g.client.Gists.Delete(g.ctx, id)

	return
}

// GetPlatformName
func (g *GistClient) GetPlatformName() string {
	return g.PlatformName
}

// GetFilterKey
func (g *GistClient) GetFilterKey() string {
	return g.FilterKey
}

// SetFilterKey
func (g *GistClient) SetFilterKey(key string) {
	g.FilterKey = key
}

// VisibilityList
func (g *GistClient) VisibilityList() (visibilityList []Visibility) {
	visibilityList = []Visibility{
		GistIsPublic,
		GistIsSecret,
	}

	return visibilityList
}

// createGistFile
func createGithubGistFiles(data []SnippetFileData) (files map[github.GistFilename]github.GistFile) {
	files = map[github.GistFilename]github.GistFile{}
	for _, d := range data {
		content := string(d.Contents)
		files[github.GistFilename(d.Path)] =
			github.GistFile{
				Content: github.String(content),
			}
	}

	return
}
