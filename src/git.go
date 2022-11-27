package src

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"path"

	"github.com/fatih/color"
	Utils "github.com/jopemachine/mac-sync-config/src/utils"
)

func GitCloneConfigsRepository() string {
	tempPath, err := os.MkdirTemp("", "mac-sync-config-temp-")
	Utils.PanicIfErr(err)

	// Should fully clone repository for commit and push
	gitCloneArgs := strings.Fields(fmt.Sprintf("git clone https://github.com/%s/%s %s", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName, tempPath))
	gitCloneCmd := exec.Command(gitCloneArgs[0], gitCloneArgs[1:]...)
	output, err := gitCloneCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)

	tempConfigDirPath := fmt.Sprintf("%s/%s", tempPath, GetRemoteConfigFolderName())

	if _, err := os.Stat(tempConfigDirPath); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(tempConfigDirPath, os.ModePerm)
		Utils.PanicIfErr(err)
	}

	return tempPath
}

func GitGetRemoteConfigHashId() string {
	gitLsRemoteArgs := strings.Fields(fmt.Sprintf("git ls-remote https://github.com/%s/%s HEAD", PreferenceSingleton.GithubId, PreferenceSingleton.MacSyncConfigGitRepositoryName))
	gitLsRemoteCmd := exec.Command(gitLsRemoteArgs[0], gitLsRemoteArgs[1:]...)
	stdout, err := gitLsRemoteCmd.CombinedOutput()
	Utils.PanicIfErr(err)

	return strings.TrimSpace(strings.Split(fmt.Sprintf("%s", stdout), "HEAD")[0])
}

func GitAddAll(cwd string) {
	gitAddAllCmd := exec.Command("git", "add", cwd)
	gitAddAllCmd.Dir = cwd
	output, err := gitAddAllCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

func GitAddFile(cwd string, filePath string) {
	gitAddFileCmd := exec.Command("git", "add", filePath)
	gitAddFileCmd.Dir = cwd
	output, err := gitAddFileCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

func GitPatchFile(cwd string, filePath string) {
	gitPatchCmd := exec.Command("git", "add", "-p", filePath)
	gitPatchCmd.Dir = cwd
	gitPatchCmd.Stdin = os.Stdin
	gitPatchCmd.Stdout = os.Stdout
	gitPatchCmd.Stderr = os.Stderr
	err := gitPatchCmd.Run()
	Utils.PanicIfErr(err)
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
	Logger.Info(fmt.Sprintf("Diff of %s\n", color.MagentaString(path.Base(filePath))))
	gitShowDiffCmd := exec.Command("git", "diff", filePath)
	gitShowDiffCmd.Dir = cwd
	gitShowDiffCmd.Stdout = os.Stdout
	gitShowDiffCmd.Stderr = os.Stderr
	gitShowDiffCmd.Run()
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