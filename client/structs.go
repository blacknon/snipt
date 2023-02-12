// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package client

// GitClient
type GitClient interface {
	// Get struct.PlatformName
	GetPlatformName() string

	// Get FilterKey
	GetFilterKey() string

	// Set FilterKey
	SetFilterKey(key string)

	// List
	List(isFile, isSecret bool) (SnippetList, error)

	// Get
	Get(id string) (SnippetData, error)

	// Create
	Create(data SnippetData) (SnippetClient, error)

	// Update
	Update(id string, data SnippetData) (SnippetClient, error)

	// Delete
	Delete(id string) error

	// VisibilityList
	VisibilityList() (visibilityList []Visibility)
}

// Snippet
type SnippetClient interface{}

// SnippetList
type SnippetList []*SnippetListData

// SnippetList.Where
func (l SnippetList) Where(fn func(*SnippetListData) bool) (result SnippetList) {
	for _, v := range l {
		if fn(v) {
			result = append(result, v)
		}
	}

	return result
}

// SnippetList
type SnippetListData struct {
	Client      GitClient
	Platform    string // platform the snippet resides on. ex) Github(hogehoge)/Gitlab(fugafuga)
	Id          string //
	Title       string //
	Description string //
	RawURL      string //
	URL         string //
	Visibility  string //
}

type SnippetData struct {
	Title       string
	Description string
	URL         string
	Visibility  Visibility
	Files       []SnippetFileData
}

func (s *SnippetData) AddFilter(val string) {
	var files []SnippetFileData
	for _, f := range s.Files {
		f.Filter = val
		files = append(files, f)
	}
	s.Files = files
}

// SnippetData
type SnippetFileData struct {
	Filter       string
	RawURL       string
	Path         string
	PreviousPath string
	Contents     []byte
}

// Visibility
type Visibility struct {
	code string
	num  int
}

func (v *Visibility) GetCode() string {
	return v.code
}

func (v *Visibility) GetNum() int {
	return v.num
}
