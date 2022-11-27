package src

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/src/utils"
)

func GitCloneConfigsRepository() string {
	tempPath, err := os.MkdirTemp("", "mac-sync-config-temp-")
	Utils.PanicIfErr(err)

	// Should fully clone repository for commit and push
	args := strings.Fields(fmt.Sprintf("git clone https://github.com/%s/%s %s", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName, tempPath))
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)

	tempConfigDirPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

	if _, err := os.Stat(tempConfigDirPath); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(tempConfigDirPath, os.ModePerm)
	}

	return tempPath
}

func GitGetRemoteConfigHashId() string {
	args := strings.Fields(fmt.Sprintf("git ls-remote https://github.com/%s/%s HEAD", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName))
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.CombinedOutput()
	Utils.PanicIfErr(err)

	return strings.TrimSpace(strings.Split(fmt.Sprintf("%s", stdout), "HEAD")[0])
}

func GitAddAll(cwd string) {
	gitAddCmd := exec.Command("git", "add", cwd)
	gitAddCmd.Dir = cwd
	output, err := gitAddCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

func GitAddFile(cwd string, filePath string) {
	gitAddCmd := exec.Command("git", "add", filePath)
	gitAddCmd.Dir = cwd
	output, err := gitAddCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

func GitPatchFile(cwd string, filePath string) {
	gitPatchCmd := exec.Command("git", "add", "-p" , filePath)
	gitPatchCmd.Dir = cwd
	gitPatchCmd.Stdout = os.Stdout
	gitPatchCmd.Stderr = os.Stderr
	err := gitPatchCmd.Run()
	Utils.PanicIfErr(err)
	Logger.NewLine()
}

func GitCommit(cwd string) {
	gitCommitCmd := exec.Command("git", "commit", "--author", "github-actions[bot] <41898282+github-actions[bot]@users.noreply.github.com>", "--allow-empty", "-m", "Commited_by_mac-sync-config")
	gitCommitCmd.Dir = cwd
	output, err := gitCommitCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

func GitPush(cwd string) {
	gitPushArgs := strings.Fields("git push -u origin main --force")
	gitPushCmd := exec.Command(gitPushArgs[0], gitPushArgs[1:]...)
	gitPushCmd.Dir = cwd
	gitPushCmd.Stdout = os.Stdout
	gitPushCmd.Stderr = os.Stderr
	err := gitPushCmd.Run()
	Utils.PanicIfErr(err)
	Logger.NewLine()
}

func GitShowDiff(cwd string, filePath string) {
	Logger.Info(fmt.Sprintf("Diff of %s\n", filePath))
	gitDiffCmd := exec.Command("git", "diff", filePath)
	gitDiffCmd.Dir = cwd
	gitDiffCmd.Stdout = os.Stdout
	gitDiffCmd.Stderr = os.Stderr
	gitDiffCmd.Run()
	// pipe might be broken.
	// Utils.PanicIfErr(err)

	Logger.NewLine()
}

func GitReset(cwd string, filePath string) {
	gitResetCmd := exec.Command("git", "checkout", "--", filePath)
	gitResetCmd.Dir = cwd
	output, err := gitResetCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

// TODO: Below command does not handle binary file properly.
func IsUpdated(cwd string, filePath string) bool {
	gitStatusCmd := exec.Command("git", "status", "-s", filePath)
	gitStatusCmd.Dir = cwd
	output, err := gitStatusCmd.CombinedOutput()
	outputStr := string(output)

	Utils.PanicIfErrWithMsg(outputStr, err)

	return len(outputStr) != 0
}
