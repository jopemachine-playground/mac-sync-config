package src

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"golang.org/x/exp/maps"
)

func executeCommand(command string, programName string) (string, error) {
	args := strings.Fields(strings.ReplaceAll(command, "{program}", programName))
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		Logger.FileLogAppend(fmt.Sprintf("%s\n%s", command, err.Error()))
		return string(output), err
	}

	Logger.FileLogAppend(fmt.Sprintf("%s\n%s", command, string(output)))
	return string(output), nil
}

func extractErrMsg(output string, err error) string {
	var errMsg string

	if strings.TrimSpace(output) == "" {
		errMsg = err.Error()
	} else {
		errMsg = output
	}

	return errMsg
}

func install(installCommand string, program string, prefix string) {
	spinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spinner.Color("cyan")
	spinner.Start()

	spinner.Suffix = fmt.Sprintf(" %s Installing \"%s\"...", prefix, program)
	output, err := executeCommand(installCommand, program)

	if err != nil {
		spinner.FinalMSG += fmt.Sprintf("%s \"%s\" failed installation\n%s\n", color.RedString("✖"), program, extractErrMsg(output, err))
	} else {
		spinner.FinalMSG += fmt.Sprintf("%s \"%s\" installed successfully\n", color.GreenString("✔"), program)
	}

	spinner.Stop()
}

func uninstall(uninstallCommand string, program string) {
	spinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spinner.Color("cyan")
	spinner.Start()

	spinner.Suffix = fmt.Sprintf(" \"%s\" is uninstalled from remote. Uninstalling \"%s\"...", program, program)
	output, err := executeCommand(uninstallCommand, program)

	if err != nil {
		spinner.FinalMSG += fmt.Sprintf("%s \"%s\" failed uninstallation\n%s\n", color.RedString("✖"), program, extractErrMsg(output, err))
	} else {
		spinner.FinalMSG += fmt.Sprintf("%s \"%s\" uninstalled successfully\n", color.GreenString("✔"), program)
	}

	spinner.Stop()
}

func SyncPrograms() {
	localDependencyCache, cacheFound := ReadDependencyCache()
	newDependencies := FetchDependencies()
	updatedDependencies := map[string]PackageManagerInfo{}
	maps.Copy(updatedDependencies, newDependencies)

	// If some local items does not exist on new dependencies, delete them.
	if cacheFound != nil {
		localDependencyCacheKeys := make([]string, 0, len(localDependencyCache))
		for k, _ := range localDependencyCache {
			localDependencyCacheKeys = append(localDependencyCacheKeys, k)
		}
		sort.Strings(localDependencyCacheKeys)

		for _, pkgManagerName := range localDependencyCacheKeys {
			pkgManagerInfo := localDependencyCache[pkgManagerName]

			// installCommand := pkgManagerInfo.InstallCommand
			uninstallCommand := pkgManagerInfo.UninstallCommand
			localInstalledPrograms := pkgManagerInfo.Programs

			// TODO: Check and handle properly case changing install command here
			// If pkgManager removed, remove all packages of the pkgManager
			if _, found := newDependencies[pkgManagerName]; !found {
				for _, program := range localInstalledPrograms {
					uninstall(uninstallCommand, program)
				}
			} else {
				for programIndex, program := range localInstalledPrograms {
					if !StringContains(newDependencies[pkgManagerName].Programs, program) {
						uninstall(uninstallCommand, program)
					} else {
						copyNewDependencies := newDependencies[pkgManagerName]
						copyNewDependencies.Programs = remove(copyNewDependencies.Programs, programIndex)
						newDependencies[pkgManagerName] = copyNewDependencies
					}
				}
			}
		}
	}

	newDependencyKeys := make([]string, 0, len(newDependencies))
	for k, _ := range newDependencies {
		newDependencyKeys = append(newDependencyKeys, k)
	}
	sort.Strings(newDependencyKeys)

	currentIdx := 0
	var totalCnt int

	for _, pkgManagerName := range newDependencyKeys {
		totalCnt += len(newDependencies[pkgManagerName].Programs)
	}

	for _, pkgManagerName := range newDependencyKeys {
		pkgManagerInfo := newDependencies[pkgManagerName]

		installCommand := pkgManagerInfo.InstallCommand
		// uninstallCommand := pkgManagerInfo.UninstallCommand
		remotePrograms := pkgManagerInfo.Programs

		for _, program := range remotePrograms {
			install(installCommand, program, fmt.Sprintf(color.WhiteString("(%d/%d)", currentIdx, totalCnt)))
			currentIdx += 1
		}
	}

	WriteDependencyCache(updatedDependencies)
}
