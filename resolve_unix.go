// +build linux darwin

package entrypoint_demoter

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func ResolveIds(match string) (uint32, uint32, error) {
	var matchStatT *syscall.Stat_t
	var uid, gid uint32
	if match != "" {
		log.Debugf("Using %s to match uid and gid", match)

		matchInfo, err := os.Stat(match)
		if err != nil {
			return 0, 0, fmt.Errorf("unable to inspect match path: %w", err)
		}

		var ok bool
		if matchStatT, ok = matchInfo.Sys().(*syscall.Stat_t); !ok {
			return 0, 0, fmt.Errorf("unsupported file info stat type: %v: %w", matchInfo.Sys(), err)
		}
	}

	uid, err := resolveIdPart("UID", matchStatT)
	if err != nil {
		return 0, 0, err
	}
	gid, err = resolveIdPart("GID", matchStatT)
	if err != nil {
		return 0, 0, err
	}

	return uid, gid, nil
}

func resolveIdPart(idPart string, matchStatT *syscall.Stat_t) (uint32, error) {
	idStr := os.Getenv(idPart)
	if idStr != "" {
		desired, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("Invalid %s: %w", idPart, err)
		}
		log.Debugf("Resolved %d from environment variable %s", desired, idPart)
		return uint32(desired), nil
	} else if matchStatT != nil {
		var desired uint32
		if idPart == "UID" {
			desired = matchStatT.Uid
		} else if idPart == "GID" {
			desired = matchStatT.Gid
		} else {
			return 0, fmt.Errorf("unknown id part: %v", idPart)
		}

		log.Debugf("Resolved %s=%d from match path", idPart, desired)
		return desired, nil
	}

	return 0, nil
}

func setCredentials(uid uint32, gid uint32, command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uid,
			Gid: gid,
		},
	}
}
