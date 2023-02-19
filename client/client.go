// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package client

import (
	"fmt"
	"os"

	"github.com/blacknon/snipt/config"
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
)

// Client
type Client struct {
	lists           []GitClient
	filterListsData SnippetList
}

// Init
func (c *Client) Init(conf config.Config) {
	// Gist.Init
	for _, gistConf := range conf.Gist {
		g := GistClient{}
		g.Init(gistConf.AccessToken)

		c.lists = append(c.lists, &g)
	}

	// Gitlab.Init
	for _, gitlabConf := range conf.GitLab {
		g := GitlabClient{}
		g.Init(gitlabConf.Url, gitlabConf.AccessToken)

		c.lists = append(c.lists, &g)
	}
}

// List
func (c *Client) List(isFile, isSecret bool) (snippetList SnippetList) {
	//
	for _, gc := range c.lists {
		list, err := gc.List(isFile, isSecret)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			continue
		}

		snippetList = append(snippetList, list...)
	}

	//
	c.filterListsData = snippetList

	return
}

// Get
func (c *Client) Get(url string) (snippet SnippetData, err error) {
	cl := c.filterListsData.Where(func(s *SnippetListData) bool {
		return s.URL == url
	})

	if len(cl) == 0 {
		return
	}

	// get SnippetListData
	sld := cl[0]

	snippet, err = sld.Client.Get(sld.Id)

	return
}

// Create
func (c *Client) Create(platform string, data SnippetData) (url []string, err error) {
	for _, d := range c.filterListsData {
		if d.Client.GetFilterKey() == platform {
			snippet, err := d.Client.Create(data)
			if err != nil {
				return url, err
			}

			switch s := snippet.(type) {
			case *github.Gist:
				url = append(url, s.GetHTMLURL())
			case *gitlab.Snippet:
				url = append(url, s.WebURL)
			}
		}
	}

	return
}

// Update
func (c *Client) Update(url string, data SnippetData) (urlList []string, err error) {
	cl := c.filterListsData.Where(func(s *SnippetListData) bool {
		return s.URL == url
	})

	for _, d := range cl {
		snippet, err := d.Client.Update(d.Id, data)
		if err != nil {
			return urlList, err
		}

		switch s := snippet.(type) {
		case *github.Gist:
			urlList = append(urlList, s.GetHTMLURL())
		case *gitlab.Snippet:
			urlList = append(urlList, s.WebURL)
		}
	}

	return
}

// Delete
func (c *Client) Delete(url string) (err error) {
	data := c.filterListsData.Where(func(s *SnippetListData) bool {
		return s.URL == url
	})

	for _, d := range data {
		err := d.Client.Delete(d.Id)
		if err != nil {
			return err
		}
	}

	return
}

// PlatformList
func (c *Client) PlatformList(enableProject bool) (platformList []string, err error) {
	// clear
	c.filterListsData = []*SnippetListData{}

	for _, gc := range c.lists {
		platformName := gc.GetPlatformName()
		gc.SetFilterKey(platformName)

		// append platform to platformList
		platformList = append(platformList, platformName)

		// append paltform to c.filterListsData
		data := &SnippetListData{
			Client:   gc,
			Platform: platformName,
		}

		c.filterListsData = append(c.filterListsData, data)

		// Get gitlab project list
		glsnippet, ok := gc.(*GitlabClient)
		if enableProject && ok {
			projects, err := glsnippet.GetProjectList()
			if err != nil {
				return []string{}, err
			}

			for _, p := range projects {
				// set platformName
				pn := fmt.Sprintf("%s /%s", platformName, p.PathWithNamespace)

				// set pd
				pd := &SnippetListData{}
				*pd = *data
				pd.Platform = platformName

				// set pd.client
				gls := &GitlabClient{}
				*gls = *glsnippet
				gls.Project = p
				gls.SetFilterKey(pn)
				pd.Client = gls

				platformList = append(platformList, pn)
				c.filterListsData = append(c.filterListsData, pd)
			}
		}
	}
	return
}

func (c *Client) VisibilityListFromPlatform(platform string) (visibilityList []Visibility) {
	for _, d := range c.filterListsData {
		if d.Client.GetFilterKey() == platform {
			visibilityList = d.Client.VisibilityList()
			break
		}
	}

	return
}

func (c *Client) VisibilityListFromURL(url string) (visibilityList []Visibility, err error) {
	cl := c.filterListsData.Where(func(s *SnippetListData) bool {
		return s.URL == url
	})

	if len(cl) == 0 {
		return
	}

	// get SnippetListData
	sld := cl[0]

	visibilityList = sld.Client.VisibilityList()

	return
}
