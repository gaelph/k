package numfmt

import (
	"os/exec"
	"strings"
)

// Formats size in human readable format
// if the -H flag is set
// uses SI if the --si flag is set
func NumFmt(num string, si bool) string {
	flag := "--to=iec"
	result := num

	if si {
		flag = "--to=si"
	}

	out, err := exec.Command("numfmt", flag, num).Output()

	if err == nil {
		result = strings.Trim(string(out), " \n\r")
	}

	return result
}
