// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package client

import (
	"fmt"
	"os"
	"sync"

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
		g := GitlabClient{
			proxy:     gitlabConf.Proxy,
			proxyUser: gitlabConf.ProxyUser,
			proxyPass: gitlabConf.ProxyPass,
		}
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
func (c *Client) PlatformList(enableProject bool) ([]string, error) {
	var platformList []string
	var err error
	var wg sync.WaitGroup // 同期用のWaitGroupを用意

	// clear
	c.filterListsData = []*SnippetListData{}

	// チャネルを用意して、処理結果を収集
	resultChannel := make(chan struct {
		platform string
		data     *SnippetListData
		err      error
	}, len(c.lists))

	// 各リストに対して並行処理を実行
	for _, gc := range c.lists {
		wg.Add(1)
		go func(gc GitClient) {
			defer wg.Done()

			platformName := gc.GetPlatformName()
			gc.SetFilterKey(platformName)

			data := &SnippetListData{
				Client:   gc,
				Platform: platformName,
			}

			// Get gitlab project list
			glsnippet, ok := gc.(*GitlabClient)
			if enableProject && ok {
				projects, err := glsnippet.GetProjectList()
				if err != nil {
					resultChannel <- struct {
						platform string
						data     *SnippetListData
						err      error
					}{platform: platformName, data: nil, err: err}
					return
				}

				for _, p := range projects {
					pn := fmt.Sprintf("%s /%s", platformName, p.PathWithNamespace)

					pd := &SnippetListData{
						Client:   &GitlabClient{Project: p /* 他のフィールドを設定 */},
						Platform: pn,
					}
					// 結果をチャネルに送信
					resultChannel <- struct {
						platform string
						data     *SnippetListData
						err      error
					}{platform: pn, data: pd, err: nil}
				}
			} else {
				// 結果をチャネルに送信
				resultChannel <- struct {
					platform string
					data     *SnippetListData
					err      error
				}{platform: platformName, data: data, err: nil}
			}
		}(gc)
	}

	// 全てのgoroutineが終了するのを待機
	go func() {
		wg.Wait()
		close(resultChannel) // チャネルを閉じる
	}()

	// 結果を受け取ってリストに追加
	for result := range resultChannel {
		if result.err != nil {
			err = result.err
		} else {
			platformList = append(platformList, result.platform)
			c.filterListsData = append(c.filterListsData, result.data)
		}
	}

	if err != nil {
		return nil, err
	}
	return platformList, nil
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
