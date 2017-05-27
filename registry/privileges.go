package registry

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

var (
	ErrUnsupportedSystem = errors.New("Unsupported system")
	ErrRootPriveleges    = errors.New("You must have root user privileges. Possibly using 'sudo' command should help")
)

// Check root rights to use system service
func checkPrivileges() error {
	if output, err := exec.Command("id", "-g").Output(); err == nil {
		if gid, parseErr := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 32); parseErr == nil {
			if gid == 0 {
				return nil
			}
			return ErrRootPriveleges
		}
	}
	return ErrUnsupportedSystem
}