package orivil

import (
	"os/exec"
	"os"
	"gopkg.in/orivil/log.v0"
	"strconv"
	"strings"
)

var SysInfo *Info

type Info struct {
	Version   string
	GoVersion string
	Process   int
	GoEnv     []string
}

func init() {
	goVersion, err := exec.Command("go", "version").Output()
	if err != nil {
		log.ErrWarnF("get go version: %v", err)
	}
	env, err := exec.Command("go", "env").Output()
	if err != nil {
		log.ErrWarnF("get go env: %v", err)
	}
	envs := strings.Split(string(env), "\n")
	SysInfo = &Info {
		Version: VERSION,
		GoVersion: string(goVersion),
		GoEnv: envs,
		Process: os.Getpid(),
	}
}

func GetSysInfo() []string {

	return SysInfo.Values()
}

func (i *Info) Values() []string {
	return append([]string{
		"orivil version: " + i.Version,
		i.GoVersion,
		"process ID: " + strconv.Itoa(i.Process),
	}, i.GoEnv...)
}
