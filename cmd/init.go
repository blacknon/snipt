// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package cmd

import (
	"path/filepath"

	"github.com/blacknon/snipt/client"
	"github.com/blacknon/snipt/config"
)

var (
	// ssss
	configFileName = "config.toml"
)

// loadConfig
func loadConfig(configFile string) (configData config.Config, err error) {
	if configFile == "" {
		dir, err := config.GetDefaultConfigDir()
		if err != nil {
			return configData, err
		}
		configFile = filepath.Join(dir, configFileName)
	}

	if err := config.Conf.Load(configFile); err != nil {
		return configData, err
	}
	configData = config.Conf

	return configData, err
}

// clientInit
func clinetInit(c string) (conf config.Config, cl client.Client, err error) {
	conf, err = loadConfig(getFullPath(c))
	if err != nil {
		return conf, cl, err
	}

	// Create client
	cl = client.Client{}
	cl.Init(conf)

	return conf, cl, err
}

// getPathList
func getPathList(args []string) (pathList []string, err error) {
	for _, a := range args {
		p := getFullPath(a)
		if isExist(p) {
			pathList = append(pathList, p)
		} else {
			return
		}
	}

	return pathList, err
}
