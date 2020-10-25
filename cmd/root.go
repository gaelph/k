/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/gaelph/k/internal/git"
	"github.com/gaelph/k/internal/numfmt"
	. "github.com/gaelph/k/internal/stat"
	"github.com/gaelph/k/internal/tabwriter"

	"github.com/spf13/cobra"

	"github.com/logrusorgru/aurora/v3"
	. "github.com/logrusorgru/aurora/v3"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

type FileDscr struct {
	name     string
	fullpath string
	fileInfo os.FileInfo
	stat     PlatformStat
}

// Return the color for a file size
func SizeToColor(size int64) uint8 {
	// 1kB
	if size <= 1024 {
		return 46
	}
	// 2kB
	if size <= 2048 {
		return 82
	}
	// 3kB
	if size <= 3072 {
		return 118
	}
	// 5kB
	if size <= 5120 {
		return 154
	}
	// 10kB
	if size <= 10240 {
		return 190
	}
	// 20kB
	if size <= 20480 {
		return 226
	}
	// 40kB
	if size <= 40960 {
		return 220
	}
	// 100kB
	if size <= 102400 {
		return 214
	}
	// 256kB
	if size <= 262144 {
		return 208
	}
	// 512kB
	if size <= 524288 {
		return 202
	}

	return 196
}

// Formats size in human readable format
// if the -H flag is set
// uses SI if the --si flag is set
func formatNumber(num int64) string {
	result := fmt.Sprint(num)

	if *humanReadableSize {

		result = numfmt.NumFmt(result, *siSize)
	}

	return result
}

// Formats and colors time
// Colors are relative to now
// TODO: accept Now as a param so that
//       all lines have the same reference
//
//           0 196  # < in the future, #spooky
//          60 255  # < less than a min old
//        3600 252  # < less than an hour old
//       86400 250  # < less than 1 day old
//      604800 244  # < less than 1 week old
//     2419200 244  # < less than 28 days (4 weeks) old
//    15724800 242  # < less than 26 weeks (6 months) old
//    31449600 240  # < less than 1 year old
//    62899200 238  # < less than 2 years old
func formatTime(t time.Time) string {
	str := t.Format("_2 Jan") + "   " + t.Format("15:04")
	secs := time.Now().Unix() - t.Unix()
	var color uint8 = 236

	if secs <= 0 {
		color = 196
	} else if secs <= 60 {
		color = 255
	} else if secs <= 3600 {
		color = 252
	} else if secs <= 86400 {
		color = 250
	} else if secs <= 604800 {
		color = 244
	} else if secs <= 2419200 {
		color = 244
	} else if secs <= 15724800 {
		color = 242
	} else if secs <= 31449600 {
		color = 240
	} else if secs <= 62899200 {
		color = 238
	}

	return aurora.Index(color, str).String()
}

// Finds the target of a symlink
func symlinkTarget(fd FileDscr) string {
	mode := fd.fileInfo.Mode()

	if mode&os.ModeSymlink == os.ModeSymlink {
		target, _ := os.Readlink(fd.fullpath)

		return " -> " + target
	}

	return ""
}

// Formats and colors a file names.
// TODO: use $LSCOLORS on macOS
// Gxfxcxdxbxegedabagacad
// case $foreground in
//     a) foreground_ansi=30;;
//     b) foreground_ansi=31;;
//     c) foreground_ansi=32;;
//     d) foreground_ansi=33;;
//     e) foreground_ansi=34;;
//     f) foreground_ansi=35;;
//     g) foreground_ansi=36;;
//     h) foreground_ansi=37;;
//     x) foreground_ansi=0;;
//   esac
func formatFilename(fd FileDscr, branch string) string {
	mode := fd.fileInfo.Mode()
	perm := mode.Perm()

	if mode.IsDir() {
		dirname := fd.name
		// writable by others
		if perm&0002 == 0002 {
			if mode&os.ModeSticky == os.ModeSticky {
				dirname = aurora.Index(0, fd.name).BgIndex(2).String()
			}
			dirname = aurora.Index(0, fd.name).BgIndex(3).String()
		}
		return dirname + " " + Gray(9, branch).String()
	}

	if mode&os.ModeSymlink == os.ModeSymlink {
		return aurora.Index(5, fd.name).BgIndex(0).String() + symlinkTarget(fd)
	}

	if mode&os.ModeSocket == os.ModeSocket {
		return aurora.Index(2, fd.name).BgIndex(0).String()
	}

	if mode&os.ModeNamedPipe == os.ModeNamedPipe {
		return aurora.Index(3, fd.name).BgIndex(0).String()
	}

	if mode&os.ModeDevice == os.ModeDevice {
		return aurora.Index(4, fd.name).BgIndex(6).String()
	}

	if mode&os.ModeCharDevice == os.ModeCharDevice {
		return aurora.Index(4, fd.name).BgIndex(3).String()
	}

	if perm&0100 == 0100 {
		if mode&os.ModeSetuid == os.ModeSetuid {
			return aurora.Index(0, fd.name).BgIndex(1).String()
		}
		if mode&os.ModeSetgid == os.ModeSetgid {
			return aurora.Index(0, fd.name).BgIndex(6).String()
		}

		return aurora.Index(1, fd.name).String()
	}

	return fd.name
}

// Returns the Git status for a file
func vcsSatus(fd FileDscr, insideVCS bool) (string, string) {
	if *noVCS {
		return "", ""
	}

	return git.Status(fd.fullpath, fd.fileInfo, insideVCS)
}

// Colors the VCS status marker
func formatVCSStatus(status string) string {
	if *noVCS {
		return ""
	}

	switch status {
	// Directory Good
	// when out of a repo, but the directory is one
	case "DG":
		return aurora.Index(46, "|").String()

		// Dirty
	case " M":
		return aurora.Index(1, "+").String()

		// Dirty+Added
	case "M ":
		return aurora.Index(82, "+").String()

		// Untracked
	case "??":
		return aurora.Index(214, "+").String()

		// Ignored
	case "!!":
		return aurora.Index(238, "|").String()

		// Added
	case "A ":
		return aurora.Index(82, "+").String()

		// Not a repo
		// when out of a repo
	case "--":
		return " "

		// Good
	default:
		return aurora.Index(82, "|").String()
	}

}

func formatUsername(username string) string {
	return Gray(9, username).String()
}

func formatGroupname(group string) string {
	return Gray(9, group).String()
}

func formatLinks(links uint16) string {
	return fmt.Sprint(links)
}

func formatSize(size int64) string {
	color := SizeToColor(size)
	str := formatNumber(size)

	return Index(color, str).String()
}

// Prints a line to a tabwrite
// with proper formating and such
func PrintLine(writer *tabwriter.Writer, f FileDscr, insideVCS bool) {
	mode := f.fileInfo.Mode().String()
	links := f.stat.Links()

	username := f.stat.Username()
	groupname := f.stat.Group()

	vcs, branch := vcsSatus(f, insideVCS)

	elemts := []string{
		mode,
		formatLinks(links),
		formatUsername(username),
		formatGroupname(groupname),
		formatSize(f.fileInfo.Size()),
		formatTime(f.fileInfo.ModTime()),
		formatVCSStatus(vcs),
		" " + formatFilename(f, branch),
	}

	fmt.Fprintln(writer, strings.Join(elemts, "\t"))
}

// Returns whether a line should be printed for a file
// accordinf to flags
func shouldPrint(name string, f os.FileInfo) bool {
	isDir := f.IsDir()
	isHidden := strings.HasPrefix(name, ".")
	showHidden := *listAlmostAll || *listAll

	should := true

	if *listDirectories && !isDir {
		should = false
	}

	if should && *dontListDirectories && isDir {
		should = false
	}

	if should && !showHidden && isHidden {
		should = false
	}

	return should
}

// Reversable sort function
func sortFn(a, b int64, reverse bool) bool {
	result := a > b

	if reverse {
		return !result
	}

	return result
}

func sortDescriptors(fds []FileDscr) []FileDscr {
	if *sortSize {
		sort.Slice(fds, func(i, j int) bool {
			sizeI := fds[i].fileInfo.Size()
			sizeJ := fds[j].fileInfo.Size()

			return sortFn(sizeI, sizeJ, *reverseSort)
		})
	}
	if *sortModTime {
		sort.Slice(fds, func(i, j int) bool {
			modTimeI := fds[i].fileInfo.ModTime().UnixNano()
			modTimeJ := fds[j].fileInfo.ModTime().UnixNano()

			return sortFn(modTimeI, modTimeJ, *reverseSort)
		})
	}
	if *sortAtime {
		sort.Slice(fds, func(i, j int) bool {
			atimeI := fds[i].stat.ATime().UnixNano()
			atimeJ := fds[j].stat.ATime().UnixNano()

			return sortFn(atimeI, atimeJ, *reverseSort)
		})
	}
	if *sortCtime {
		sort.Slice(fds, func(i, j int) bool {
			ctimeI := fds[i].stat.CTime().UnixNano()
			ctimeJ := fds[j].stat.CTime().UnixNano()

			return sortFn(ctimeI, ctimeJ, *reverseSort)
		})
	}

	return fds
}

func handleSortFlag(cmd *cobra.Command) {
	sortBy, _ = cmd.Flags().GetString("sort")

	for _, r := range sortBy {
		switch r {
		case 's':
			*sortSize = true
			continue

		case 't':
			*sortModTime = true
			continue

		case 'c':
			*sortCtime = true
			continue

		case 'a':
			*sortAtime = true
			continue
		}
	}
}

func handleArgs(args []string) string {
	cwd, _ := os.Getwd()

	if len(args) == 1 {
		target := args[0]

		cwd = path.Clean(path.Join(cwd, target))
	}

	if err := os.Chdir(cwd); err != nil {
		panic(err)
	}

	return cwd
}

func getDescriptors(cwd string) []FileDscr {
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		panic(err)
	}

	descriptors := make([]FileDscr, 0)

	// Add . and ..
	if *listAll {
		dot, _ := os.Stat(cwd)
		dotdot, _ := os.Stat(path.Dir(cwd))

		dotD := FileDscr{
			".",
			cwd,
			dot,
			NewPlatformStat(dot),
		}

		dotdotD := FileDscr{
			"..",
			path.Dir(cwd),
			dotdot,
			NewPlatformStat(dotdot),
		}

		descriptors = append(descriptors, dotD, dotdotD)
	}

	// Actual file list
	for _, file := range files {
		if shouldPrint(file.Name(), file) {
			descriptors = append(descriptors, FileDscr{
				file.Name(),
				path.Join(cwd, file.Name()),
				file,
				NewPlatformStat(file),
			})
		}
	}

	return sortDescriptors(descriptors)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k [directory]",
	Short: "k makes directory listings more readable",
	Long: `k makes directory listings more readable,
adding a bit of color and some git status information
on files and directories.`,
	// no args == current dir
	// 1 arg == dir relative to the current one
	// TODO: handle absolute paths
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		// Waiting line
		fmt.Print("Reading directory…")

		handleSortFlag(cmd)

		cwd := handleArgs(args)
		descriptors := getDescriptors(cwd)

		insideVCS := git.IsInWorkTree()

		writer := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', tabwriter.AlignRight)

		var blocks int64 = 0
		for _, d := range descriptors {
			blocks += d.stat.Blocks()

			PrintLine(writer, d, insideVCS)
		}

		// Clear waiting line
		fmt.Print("\r                  \r")

		// Final Output
		fmt.Printf(" total %d\n", blocks)
		writer.Flush()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var listAll *bool
var listAlmostAll *bool
var sortCtime *bool
var listDirectories *bool
var dontListDirectories *bool
var humanReadableSize *bool
var siSize *bool
var reverseSort *bool
var sortSize *bool
var sortModTime *bool
var sortAtime *bool
var dontSort *bool
var sortBy string
var noVCS *bool

func init() {
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	listAll = rootCmd.Flags().BoolP("all", "a", false, "list entries starting .")
	listAlmostAll = rootCmd.Flags().BoolP("almost-all", "A", false, "list all except . and ..")
	sortCtime = rootCmd.Flags().BoolP("ctime", "c", false, "sort by ctime")
	listDirectories = rootCmd.Flags().BoolP("directories", "d", false, "list only directories")
	dontListDirectories = rootCmd.Flags().BoolP("no-directories", "n", false, "do not list directories")
	humanReadableSize = rootCmd.Flags().BoolP("human", "H", false, "show file sizes in human readable format")
	siSize = rootCmd.Flags().Bool("si", false, "with -h, use powers of 1000 not 1024")
	reverseSort = rootCmd.Flags().BoolP("reverse", "r", false, "reverse sort order")
	sortSize = rootCmd.Flags().BoolP("size", "S", false, "sort by size")
	sortModTime = rootCmd.Flags().BoolP("time", "t", false, "sort by modification time")
	sortAtime = rootCmd.Flags().BoolP("atime", "u", false, "sort by atime (use of access time)")
	dontSort = rootCmd.Flags().BoolP("unsorted", "U", false, "unsorted")

	rootCmd.Flags().String("sort", "n", "sort by WORD: none (U), size (s),\ntime (t), ctime or status (c),\natime or access time or use (a)")

	noVCS = rootCmd.Flags().Bool("no-vcs", false, "do not get VCS stats (much faster)")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".k" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".k")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
