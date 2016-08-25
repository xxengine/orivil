// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

import (
	"gopkg.in/orivil/config.v0"
	"gopkg.in/orivil/helper.v0"
	"gopkg.in/orivil/session.v0"
	"os"
	"os/exec"
	"path/filepath"
	"gopkg.in/orivil/log.v0"
	"fmt"
)

// server config
var CfgApp = &struct {
	DEBUG                     bool
	KEY                       string
	VIEW_FILE_EXT             string
	MEMORY_SESSION_KEY        string
	MEMORY_SESSION_MAX_AGE    int // minute
	MEMORY_GC_CHECK_NUM       int
	PERMANENT_SESSION_KEY     string
	PERMANENT_SESSION_MAX_AGE int // minute
	PERMANENT_GC_CHECK_NUM    int
	READ_TIMEOUT              int // second
	WRITE_TIMEOUT             int // second
}{
	// default config
	DEBUG:                     true,
	KEY:                       "--------------------------------------",
	VIEW_FILE_EXT:             ".tmpl",
	MEMORY_SESSION_KEY:        "orivil-memory-session",
	MEMORY_SESSION_MAX_AGE:    45,
	MEMORY_GC_CHECK_NUM:       3,
	PERMANENT_SESSION_KEY:     "orivil-permanent-session",
	PERMANENT_SESSION_MAX_AGE: 45,
	PERMANENT_GC_CHECK_NUM:    3,
	READ_TIMEOUT:              30,
	WRITE_TIMEOUT:             30,
}

// dirs
var (
	DirBase = getBaseDir()
	DirStaticFile = filepath.Join(DirBase, "public")
	DirBundle = filepath.Join(DirBase, "bundle")
	DirConfig = filepath.Join(DirBase, "config")
	DirCache = filepath.Join(DirBase, "cache")
)

var Cfg = config.NewConfig(DirConfig)

func init() {

	err := Cfg.ReadStruct("app.yml", CfgApp)
	if err != nil {
		log.ErrInfoF("%v\nread 'app.yml' got error, use default value instead.", err)
	}

	Key = CfgApp.KEY

	session.ConfigMemory(
		CfgApp.MEMORY_SESSION_MAX_AGE,
		CfgApp.MEMORY_GC_CHECK_NUM,
		CfgApp.MEMORY_SESSION_KEY,
	)
	session.ConfigPermanent(
		CfgApp.PERMANENT_SESSION_MAX_AGE,
		CfgApp.PERMANENT_GC_CHECK_NUM,
		CfgApp.PERMANENT_SESSION_KEY,
	)
}

func getBaseDir() string {
	cDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if helper.IsExist(filepath.Join(cDir, "bundle")) {
		return cDir
	} else {
		file, err := exec.LookPath(os.Args[0])
		if err != nil {
			panic(err)
		}
		path, err := filepath.Abs(file)
		if err != nil {
			panic(err)
		}
		eDir := filepath.Dir(path)

		if helper.IsExist(filepath.Join(eDir, "bundle")) {
			return eDir
		} else {
			fmt.Printf("Directory 'bundle' not exist in current directory: [%s] "+
			"or executable file directory: [%s]", cDir, eDir)
			os.Exit(1)
		}
	}
	return ""
}
