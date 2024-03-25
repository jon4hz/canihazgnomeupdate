package extensions

import (
	"os/exec"
	"strings"
)

func List(enabled bool) ([]string, error) {
	cmdArgs := []string{"list"}
	if enabled {
		cmdArgs = append(cmdArgs, "--enabled")
	}

	// Run the command
	cmd := exec.Command("gnome-extensions", cmdArgs...)
	cmdOut, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse the output
	extensions := []string{}
	for _, line := range strings.Split(string(cmdOut), "\n") {
		if line != "" {
			extensions = append(extensions, line)
		}
	}

	return extensions, nil
}
