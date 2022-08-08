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
	if args[0] == "sudo" {
		cmd.Stdin = strings.NewReader(PreferenceSingleton.UserPassword)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		Logger.FileLogAppend(fmt.Sprintf("✖ %s\n%s\n", command, err.Error()))
		return string(output), err
	}

	Logger.FileLogAppend(fmt.Sprintf("ℹ %s\n%s", command, string(output)))
	return string(output), nil
}

func extractErrMsg(output string, err error) string {
	if strings.TrimSpace(output) == "" {
		return err.Error()
	} else {
		return output
	}
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
	localProgramCache := ReadLocalProgramCache()
	newPrograms := FetchDependencies()
	updatedPrograms := map[string]PackageManagerInfo{}
	maps.Copy(updatedPrograms, newPrograms)

	// If some local items does not exist on new dependencies, delete them.
	if localProgramCache != nil {
		localProgramCacheKeys := make([]string, 0, len(localProgramCache))
		for k, _ := range localProgramCache {
			localProgramCacheKeys = append(localProgramCacheKeys, k)
		}
		sort.Strings(localProgramCacheKeys)

		for _, pkgManagerName := range localProgramCacheKeys {
			pkgManagerInfo := localProgramCache[pkgManagerName]

			// installCommand := pkgManagerInfo.InstallCommand
			uninstallCommand := pkgManagerInfo.UninstallCommand
			localInstalledPrograms := pkgManagerInfo.Programs

			// If pkgManager removed, remove all packages of the pkgManager
			if _, found := newPrograms[pkgManagerName]; !found {
				for _, program := range localInstalledPrograms {
					uninstall(uninstallCommand, program)
				}
			} else {
				for programIndex, program := range localInstalledPrograms {
					if !StringContains(newPrograms[pkgManagerName].Programs, program) {
						uninstall(uninstallCommand, program)
					} else {
						copynewPrograms := newPrograms[pkgManagerName]
						copynewPrograms.Programs = Remove(copynewPrograms.Programs, programIndex)
						newPrograms[pkgManagerName] = copynewPrograms
					}
				}
			}
		}
	}

	newProgramsKeys := make([]string, 0, len(newPrograms))
	for k, _ := range newPrograms {
		newProgramsKeys = append(newProgramsKeys, k)
	}
	sort.Strings(newProgramsKeys)

	currentIdx := 0
	var totalCnt int

	for _, pkgManagerName := range newProgramsKeys {
		totalCnt += len(newPrograms[pkgManagerName].Programs)
	}

	for _, pkgManagerName := range newProgramsKeys {
		pkgManagerInfo := newPrograms[pkgManagerName]

		installCommand := pkgManagerInfo.InstallCommand
		// uninstallCommand := pkgManagerInfo.UninstallCommand
		remotePrograms := pkgManagerInfo.Programs

		for _, program := range remotePrograms {
			install(installCommand, program, fmt.Sprintf(color.WhiteString("(%d/%d)", currentIdx, totalCnt)))
			currentIdx += 1
		}
	}

	if totalCnt == 0 {
		Logger.Success("All programs up to dated.")
	} else {
		Logger.Success(fmt.Sprintf("%d programs updated.", totalCnt))
	}

	WriteLocalProgramCache(updatedPrograms)
	Logger.WriteFileLog("./logs")
}
