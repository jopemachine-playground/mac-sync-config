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

func ReadDependencyCache() (map[string]PackageManagerInfo, error) {
	if _, err := os.Stat(DependencyCachePath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	dat, err := ioutil.ReadFile(DependencyCachePath)
	if err != nil {
		panic(err)
	}

	var config map[string]PackageManagerInfo

	if err := yaml.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	return config, nil
}

func ReadConfigCache() (ConfigInfo, error) {
	if _, err := os.Stat(ConfigCachePath); errors.Is(err, os.ErrNotExist) {
		return ConfigInfo{}, err
	}

	dat, err := ioutil.ReadFile(ConfigCachePath)

	if err != nil {
		panic(err)
	}

	var config ConfigInfo

	if err := yaml.Unmarshal(dat, &config); err != nil {
		panic(err)
	}

	return config, nil
}

func WriteDependencyCache(cache map[string]PackageManagerInfo) {
	bytesToWrite, err := GetBytes(cache)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(DependencyCachePath); !errors.Is(err, os.ErrNotExist) {
		os.Remove(DependencyCachePath)
	}

	if writeErr := ioutil.WriteFile(DependencyCachePath, bytesToWrite, 0644); writeErr != nil {
		panic(writeErr)
	}
}
