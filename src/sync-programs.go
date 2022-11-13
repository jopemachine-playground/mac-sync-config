package src

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func executeCommand(command string, programName string) error {
	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)

	// if args[0] == "sudo" {
	//	cmd.Stdin = strings.NewReader(PreferenceSingleton.UserPassword)
	// }

	cmd.Stdin = strings.NewReader(PreferenceSingleton.UserPassword)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractErrMsg(output string, err error) string {
	if strings.TrimSpace(output) == "" {
		return err.Error()
	} else {
		return output
	}
}

func install(command string, program string, progress string) error {
	command = strings.ReplaceAll(command, "{program}", program)
	Logger.Log(fmt.Sprintf("%s Installing \"%s\"... %s", fmt.Sprint(progress), program, fmt.Sprintf("(%s)", command)))

	err := executeCommand(command, program)
	Logger.Log("")
	return err
}

func uninstall(command string, program string) error {
	command = strings.ReplaceAll(command, "{program}", program)
	Logger.Log(fmt.Sprintf("\"%s\" is uninstalled from remote. Uninstalling \"%s\"...", program, program))

	err := executeCommand(command, program)
	Logger.Log("")
	return err
}

func SyncPrograms() {
	var totalCnt int
	var failedCnt int

	localProgramCache := ReadLocalProgramCache()
	newPrograms := FetchRemoteProgramInfo()

	// If some local items does not exist on new dependencies, delete them.
	if localProgramCache != nil {
		localProgramCacheKeys := make([]string, 0, len(localProgramCache))
		for k := range localProgramCache {
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
					if err := uninstall(uninstallCommand, program); err != nil {
						failedCnt += 1
					}
				}
				totalCnt += len(localInstalledPrograms)
			} else {
				for _, program := range localInstalledPrograms {
					if !StringContains(newPrograms[pkgManagerName].Programs, program) {
						if err := uninstall(uninstallCommand, program); err != nil {
							uninstall(uninstallCommand, program)
							failedCnt += 1
						}
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
	for k := range newPrograms {
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
			progress := fmt.Sprintf("[%d/%d]", currentIdx, totalCnt)

			if err := install(installCommand, program, progress); err != nil {
				failedCnt += 1
			}
			currentIdx += 1
		}
	}

	if totalCnt == 0 {
		Logger.Success("All programs are already synced.")
	} else {
		Logger.Success(fmt.Sprintf("%d items updated successfully, %d items update failed.", totalCnt, failedCnt))
	}

	// TODO: Remove below request, deep copy the info object.
	WriteLocalProgramCache(FetchRemoteProgramInfo())
}
