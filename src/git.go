package src

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	Utils "github.com/jopemachine/mac-sync-config/utils"
)

type gitManipulator struct{}

var (
	Git gitManipulator
)

func (git gitManipulator) AddAllFiles(cwd string) {
	gitAddAllCmd := exec.Command("git", "add", cwd)
	gitAddAllCmd.Dir = cwd
	output, err := gitAddAllCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

func (git gitManipulator) AddFile(cwd string, filePath string) {
	gitAddFileCmd := exec.Command("git", "add", filePath)
	gitAddFileCmd.Dir = cwd
	output, err := gitAddFileCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

func (git gitManipulator) PatchFile(cwd string, filePath string) {
	gitPatchCmd := exec.Command("git", "add", "-p", filePath)
	gitPatchCmd.Dir = cwd
	gitPatchCmd.Stdin = os.Stdin
	gitPatchCmd.Stdout = os.Stdout
	gitPatchCmd.Stderr = os.Stderr
	Utils.FatalExitIfError(gitPatchCmd.Run())
}

func (git gitManipulator) Commit(cwd string) {
	Logger.Info("Enter the commit message.")
	gitCommitCmd := exec.Command("git", "commit", "--author", GH_BOT_EMAIL, "--allow-empty")
	gitCommitCmd.Dir = cwd
	gitCommitCmd.Env = append(gitCommitCmd.Env, "GIT_COMMITTER_NAME=\"Mac-sync-config\"")
	gitCommitCmd.Env = append(gitCommitCmd.Env, "EDITOR=vim")
	gitCommitCmd.Env = append(gitCommitCmd.Env, fmt.Sprintf("TERM=%s", os.Getenv("TERM")))
	// TODO: Fix backspace key not working in the vim issue
	gitCommitCmd.Stdin = os.Stdin
	gitCommitCmd.Stdout = os.Stdout
	gitCommitCmd.Stderr = os.Stderr

	// Assume the error is caused by user's abort.
	if err := gitCommitCmd.Run(); err != nil {
		Logger.Error("Git commit aborted.")
		os.Exit(1)
	}

	Logger.ClearConsole()
}

func (git gitManipulator) Push(cwd string) {
	gitPushArgs := strings.Fields(fmt.Sprintf("git push -u origin %s --force", GetGitBranchName()))
	gitPushCmd := exec.Command(gitPushArgs[0], gitPushArgs[1:]...)
	gitPushCmd.Dir = cwd
	gitPushCmd.Stdout = os.Stdout
	gitPushCmd.Stderr = os.Stderr
	Utils.FatalExitIfError(gitPushCmd.Run())
	Logger.NewLine()
}

func (git gitManipulator) ShowDiff(cwd string, filePath string) {
	// TODO: It would be good to check only once.
	checkDiffCmd := exec.Command("git", "diff", "--quiet", filePath)
	checkDiffCmd.Dir = cwd
	checkDiffCmd.Stdout = os.Stdout
	checkDiffCmd.Stderr = os.Stderr

	if err := checkDiffCmd.Run(); err == nil {
		Logger.Info("No diff to show")
	} else {
		gitShowDiffCmd := exec.Command("git", "diff", filePath)
		gitShowDiffCmd.Dir = cwd
		gitShowDiffCmd.Stdout = os.Stdout
		gitShowDiffCmd.Stderr = os.Stderr
		gitShowDiffCmd.Run()

		// pipe might be broken, but maybe doesn't matter here.
		// Utils.FatalExitIfError(err)

		Logger.NewLine()
	}
}

func (git gitManipulator) Reset(cwd string, filePath string) {
	gitResetCmd := exec.Command("git", "checkout", "--", filePath)
	gitResetCmd.Dir = cwd
	output, err := gitResetCmd.CombinedOutput()
	Utils.PanicIfErrWithMsg(string(output), err)
}

// TODO: Below command does not handle binary file properly.
func (git gitManipulator) IsUpdated(cwd string, filePath string) bool {
	gitStatusCmd := exec.Command("git", "status", "-s", filePath)
	gitStatusCmd.Dir = cwd
	output, err := gitStatusCmd.CombinedOutput()
	outputStr := string(output)

	Utils.PanicIfErrWithMsg(outputStr, err)

	return len(outputStr) != 0
}
