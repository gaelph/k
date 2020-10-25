package git

// Helper function for git status

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func trimAllSpaces(s string) string {
	return strings.Trim(s, " \n\r\t\v\f")
}

// Returns true if th current working directory
// is in a git work tree
func IsInWorkTree() bool {
	var err error

	s, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()

	if err != nil {
		return false
	}

	return trimAllSpaces(string(s)) == "true"
}

// TopLevel tries to get the top level direcory of
// the current git working directory
//
// If 'filepath' is a file, returns a bool indicating
// whether it is in a git work tree
// If 'filepath' is a directory, returns a bool indicatind
// whether it is a git repository
//
// The second return value is the path to the top level
// git repository (if the first return value is true
func TopLevel(filepath string, isDir bool) (bool, string) {
	var err error
	cwd, err := os.Getwd()
	defer func() {
		if isDir {
			os.Chdir(cwd)
		}
	}()

	if err != nil {
		panic(err)
	}

	inVCS := false
	topLevel := ""

	if isDir {
		err = os.Chdir(filepath)
	}

	if err == nil {
		inVCS = IsInWorkTree()

		if inVCS {
			var t []byte
			t, err = exec.Command("git", "rev-parse", "--show-toplevel").Output()

			topLevel = trimAllSpaces(string(t))
		}
	}

	if err != nil {
		panic(err)
	}

	return inVCS, topLevel
}

// Gets the git branch name
// 'toplevel' is the path the top level git repository
// 'dir' is the path to the directory we want to check
// Returns an empty string if dir is not a repository
func GetBranchName(topLevel string, dir string) string {
	b, err := exec.Command("git", fmt.Sprintf("--git-dir=%s/.git", topLevel), fmt.Sprintf("--work-tree=%s", dir), "rev-parse", "--abbrev-ref", "HEAD").Output()

	if err != nil {
		return ""
	}

	return strings.Trim(string(b), " \n\r\t\v\f")
}

// Get the git status for a repository
// 'topLevel' is the path to the top level repository
// 'dir' is the path to the directory we want to check
//
// If 'dir' is not a repository, return "--"
// If 'dir' has no pending changes, returns "DG"
// If 'dir' is dirty, return " M"
func RepoStatus(topLevel string, dir string) string {
	var status string = "--"
	s := exec.Command("git", fmt.Sprintf("--git-dir=%s/.git", topLevel), fmt.Sprintf("--work-tree=%s", dir), "diff", "--stat", "--quiet", "--ignore-submodules", "HEAD").ProcessState.ExitCode()

	if s == 1 {
		// Custom status "Directory Good"
		status = "DG"
	} else {
		status = " M"
	}

	return status
}

// Returns true if 'path' is gitignored
func IsIgnored(path string) bool {
	c := exec.Command("git", "check-ignore", "--quiet", path)
	c.Run()

	s := c.ProcessState.ExitCode()

	return s == 0
}

// Returns true if directroy at 'dirpath' has changes
func HasDirectoryChanges(dirpath string) bool {
	c := exec.Command("git", "diff", "--stat", "--exit-code", "--quiet", "--ignore-submodules", dirpath)
	c.Run()

	s := c.ProcessState.ExitCode()

	return s == 1
}

// Returns status of a directory inside a git repo
// "  " -> no changes
// "!!" -> ignored
// " M" -> pending changes
func DirectoryStatus(dir string) string {
	var status string = "  "

	if IsIgnored(dir) {
		status = "!!"
	} else if HasDirectoryChanges(dir) {
		status = " M"
	}

	return status
}

// Returns status of a file inside a git repo
func FileStatus(path string) string {
	var status string = "  "
	o, err := exec.Command("git", "status", "--porcelain", "--ignored", "--untracked-files=normal", path).Output()

	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(o), "\n")

	if len(lines) > 0 {

		out := strings.Trim(lines[0], "\n\r\t\v\f")

		if len(out) >= 2 {
			status = out[0:2]
		}
	}

	return status
}

// Returns the status of a file/directory/repo
// If fullpath is a repository, the second return value is
// the git brach this repository is on
func Status(fullpath string, file os.FileInfo, insideVCS bool) (string, string) {

	status := "--" // Custom status for "not a repo"
	branch := ""
	isDir := file.IsDir()

	isRepo, topLevel := TopLevel(fullpath, isDir)

	if isRepo {
		// The file is a repository, but we are not in one
		// Display the branch and status (good/dirty)
		if !insideVCS {
			branch = GetBranchName(topLevel, fullpath)

			if isDir {
				status = RepoStatus(topLevel, fullpath)
			}
		} else {
			if isDir {
				status = DirectoryStatus(fullpath)
			} else {
				status = FileStatus(fullpath)
			}
		}
	}

	return status, branch
}
