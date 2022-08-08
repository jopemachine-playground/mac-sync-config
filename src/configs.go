package src

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const CachePath = "~/Library/Caches/Mac-sync"

var (
	ProgramCachePath = strings.Join([]string{CachePath, "local-programs.yaml"}, "/")
)

type PackageManagerInfo struct {
	InstallCommand   string   `yaml:"install"`
	UninstallCommand string   `yaml:"uninstall"`
	Programs         []string `yaml:"programs"`
}

type ConfigInfo struct {
	ConfigPathsToSync []string `yaml:"sync"`
}

func ReadLocalProgramCache() map[string]PackageManagerInfo {
	programCachePath := HandleTildePath(ProgramCachePath)

	if _, err := os.Stat(programCachePath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	dat, err := ioutil.ReadFile(programCachePath)
	if err != nil {
		panic(err)
	}

	var config map[string]PackageManagerInfo

	if err := yaml.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	return config
}

func WriteLocalProgramCache(cache map[string]PackageManagerInfo) {
	cacheDir := HandleTildePath(CachePath)
	programCachePath := HandleTildePath(ProgramCachePath)

	if _, err := os.Stat(cacheDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(cacheDir, 0777)
		if err != nil {
			panic(err)
		}
	} else if _, err := os.Stat(programCachePath); !errors.Is(err, os.ErrNotExist) {
		os.Remove(programCachePath)
	}

	bytesToWrite, err := yaml.Marshal(cache)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(programCachePath, bytesToWrite, 0777); err != nil {
		panic(err)
	}
}

func ClearCache() {
	programCachePath := HandleTildePath(ProgramCachePath)
	os.Remove(programCachePath)
}
