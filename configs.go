package orivil

import (
	"gopkg.in/orivil/config.v0"
	"gopkg.in/orivil/helper.v0"
	"gopkg.in/orivil/session.v0"
	"os"
	"os/exec"
	"path/filepath"
)

// server config
var CfgApp = &struct {
	Debug                     bool
	Key                       string
	View_file_ext             string
	Memory_session_key        string
	Memory_session_max_age    int // minute
	Memory_GC_check_num       int
	Permanent_session_key     string
	Permanent_session_max_age int // minute
	Permanent_GC_check_num    int
	Timeout                   int // second
}{
	// default config
	Debug:                     true,
	Key:                       "jfsjlfmklsiejojwio8392ufk-fc0sjlmp;skf[wfjoshu",
	View_file_ext:             ".tmpl",
	Memory_session_key:        "orivil-memory-session",
	Permanent_session_key:     "orivil-permanent-session",
	Memory_GC_check_num:       3,
	Memory_session_max_age:    45,
	Permanent_session_max_age: 45,
	Permanent_GC_check_num:    3,
	Timeout:                   10,
}

// dirs
var (
	DirBase       = getBaseDir()
	DirStaticFile = filepath.Join(DirBase, "public")
	DirBundle     = filepath.Join(DirBase, "bundle")
	DirConfig     = filepath.Join(DirBase, "config")
	DirCache      = filepath.Join(DirBase, "cache")
)

var Cfg = config.NewConfig(DirConfig)

func init() {

	Cfg.ReadStruct("app.yml", CfgApp)

	Key = CfgApp.Key

	session.ConfigMemory(
		CfgApp.Memory_session_max_age,
		CfgApp.Memory_GC_check_num,
		CfgApp.Memory_session_key,
	)
	session.ConfigPermanent(
		CfgApp.Permanent_session_max_age,
		CfgApp.Permanent_GC_check_num,
		CfgApp.Permanent_session_key,
	)
}

func getBaseDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if helper.IsExist(filepath.Join(dir, "bundle")) {
		return dir
	} else {
		file, err := exec.LookPath(os.Args[0])
		if err != nil {
			panic(err)
		}
		path, err := filepath.Abs(file)
		if err != nil {
			panic(err)
		}
		dir = filepath.Dir(path)
		return dir
	}
}
