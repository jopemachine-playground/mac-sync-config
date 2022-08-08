package src

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/fatih/color"
)

func executeCommand(command string, programName string) error {
	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)

	if args[0] == "sudo" {
		cmd.Stdin = strings.NewReader(PreferenceSingleton.UserPassword)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

func extractErrMsg(output string, err error) string {
	if strings.TrimSpace(output) == "" {
		return err.Error()
	} else {
		return output
	}
}

func install(command string, program string, progress string) {
	command = strings.ReplaceAll(command, "{program}", program)
	Logger.Log(color.New(color.FgWhite, color.Bold).Sprintf("%s Installing \"%s\"... %s", GrayColor.Sprint(progress), program, GrayColor.Sprintf("(%s)", command)))

	executeCommand(command, program)
	Logger.Log("")
}

func uninstall(command string, program string) {
	command = strings.ReplaceAll(command, "{program}", program)
	Logger.Log(color.New(color.FgWhite, color.Bold).Sprintf("\"%s\" is uninstalled from remote. Uninstalling \"%s\"...", program, program))

	executeCommand(command, program)
	Logger.Log("")
}

func SyncPrograms() {
	var totalCnt int

	localProgramCache := ReadLocalProgramCache()
	newPrograms := FetchRemoveProgramInfo()

	// If some local items does not exist on new dependencies, delete them.
	if localProgramCache != nil {
		localProgramCacheKeys := make([]string, 0, len(localProgramCache))
		for k, _ := range localProgramCache {
			localProgramCacheKeys = append(localProgramCacheKeys, k)
		}
		sort.Strings(localProgramCacheKeys)

		for _, pkgManagerName := range localProgramCacheKeys {
			pkgManagerInfo := localProgramCache[pkgManagerName]

			uninstallCommand := pkgManagerInfo.UninstallCommand
			localInstalledPrograms := pkgManagerInfo.Programs

			// If pkgManager removed, remove all packages of the pkgManager
			if _, found := newPrograms[pkgManagerName]; !found {
				for _, program := range localInstalledPrograms {
					uninstall(uninstallCommand, program)
				}
				totalCnt += len(localInstalledPrograms)
			} else {
				for _, program := range localInstalledPrograms {
					if !StringContains(newPrograms[pkgManagerName].Programs, program) {
						uninstall(uninstallCommand, program)
						totalCnt += 1
					} else {
						newProgramsCopy := newPrograms[pkgManagerName]
						for programIndex, _ := range newProgramsCopy.Programs {
							if newProgramsCopy.Programs[programIndex] == program {
								newProgramsCopy.Programs = Remove(newProgramsCopy.Programs, programIndex)
								break
							}
						}
						newPrograms[pkgManagerName] = newProgramsCopy
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

	for _, pkgManagerName := range newProgramsKeys {
		totalCnt += len(newPrograms[pkgManagerName].Programs)
	}

	for _, pkgManagerName := range newProgramsKeys {
		pkgManagerInfo := newPrograms[pkgManagerName]

		installCommand := pkgManagerInfo.InstallCommand
		remotePrograms := pkgManagerInfo.Programs

		for _, program := range remotePrograms {
			install(installCommand, program, fmt.Sprintf("[%d/%d]", currentIdx, totalCnt))
			currentIdx += 1
		}
	}

	if totalCnt == 0 {
		Logger.Success("All programs are already synced")
	} else {
		Logger.Success(fmt.Sprintf("%d programs updated.", totalCnt))
	}

	// TODO: Remove below request, deep copy the info object.
	WriteLocalProgramCache(FetchRemoveProgramInfo())
}
