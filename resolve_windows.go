// +build windows

package entrypoint_demoter

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
)

func ResolveIds(match string) (uint32, uint32, error) {
	var uid, gid uint32
	if match != "" {
		log.Debugf("Would use %s to match uid and gid", match)
	}

	uid, err := resolveIdPart("UID")
	if err != nil {
		return 0, 0, err
	}
	gid, err = resolveIdPart("GID")
	if err != nil {
		return 0, 0, err
	}

	return uid, gid, nil
}

func resolveIdPart(idPart string) (uint32, error) {
	idStr := os.Getenv(idPart)
	if idStr != "" {
		desired, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", idPart, err)
		}
		log.Debugf("Resolved %d from environment variable %s", desired, idPart)
		return uint32(desired), nil
	}

	return 0, nil
}

//noinspection GoUnusedParameter
func setCredentials(uid uint32, gid uint32, command *exec.Cmd) {
	// do nothing, just log
	log.
		WithField("uid", uid).
		WithField("gid", gid).
		Warn("Running on Windows, so not setting command credentials")
}
