package src

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CloneMacSyncConfigRepository() string {
	tempPath, err := os.MkdirTemp("", "mac-sync-config-temp-")
	PanicIfErr(err)

	// Should fully clone repository for commit and push
	args := strings.Fields(fmt.Sprintf("git clone https://github.com/%s/%s %s", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName, tempPath))
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)

	tempConfigDirPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

	if _, err := os.Stat(tempConfigDirPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempConfigDirPath, os.ModePerm)
	}

	return tempPath
}

func GitAddCwd(cwd string) {
	gitAddCmd := exec.Command("git", "add", cwd)
	gitAddCmd.Dir = cwd
	output, err := gitAddCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)
}

func GitAddFile(cwd string, filePath string) {
	gitAddCmd := exec.Command("git", "add", filePath)
	gitAddCmd.Dir = cwd
	output, err := gitAddCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)
}

func GitCommit(cwd string) {
	gitCommitCmd := exec.Command("git", "commit", "--author", "github-actions[bot] <41898282+github-actions[bot]@users.noreply.github.com>", "--allow-empty", "-m", "Commited_by_mac-sync-config")
	gitCommitCmd.Dir = cwd

	output, err := gitCommitCmd.CombinedOutput()
	PanicIfErrWithOutput(string(output), err)
}

func GitPush(cwd string) {
	gitPushArgs := strings.Fields("git push -u origin main --force")
	gitPushCmd := exec.Command(gitPushArgs[0], gitPushArgs[1:]...)
	gitPushCmd.Dir = cwd
	gitPushCmd.Stdout = os.Stdout
	gitPushCmd.Stderr = os.Stderr
	err := gitPushCmd.Run()
	PanicIfErr(err)
}

func IsUpdated(cwd string, filePath string) bool {
	gitDiffCmd := exec.Command("git", "diff", filePath)
	gitDiffCmd.Dir = cwd
	output, err := gitDiffCmd.CombinedOutput()
	outputStr := string(output)

	PanicIfErrWithOutput(outputStr, err)

	if len(outputStr) == 0 {
		return false
	}

	return true
}

func ShowDiff(cwd string, filePath string) {
	gitDiffCmd := exec.Command("git", "diff", filePath)
	gitDiffCmd.Dir = cwd

	gitDiffCmd.Stdout = os.Stdout
	gitDiffCmd.Stderr = os.Stderr
	err := gitDiffCmd.Run()
	PanicIfErr(err)
}
