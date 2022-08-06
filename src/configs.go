package src

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
)

const CachePath = "~/Library/Caches/Mac-sync"

var (
	DependencyCachePath = strings.Join([]string{CachePath, "local-dependency.yaml"}, "/")
	ConfigCachePath     = strings.Join([]string{CachePath, "local-configs.yaml"}, "/")
)

type PackageManagerInfo struct {
	InstallCommand   string   `yaml:"install"`
	UninstallCommand string   `yaml:"uninstall"`
	Programs         []string `yaml:"programs"`
}

type ConfigInfo struct {
	ConfigPathsToSync []string `yaml:"sync"`
}

func ReadDependencyCache() interface{} {
	dat, err := ioutil.ReadFile(DependencyCachePath)
	if err != err {
		panic(err)
	}

	var config map[string]PackageManagerInfo

	if err := yaml.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	return config
}

func ReadConfigCache() interface{} {
	dat, err := ioutil.ReadFile(ConfigCachePath)
	if err != err {
		panic(err)
	}

	var config ConfigInfo

	if err := yaml.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	return config
}
