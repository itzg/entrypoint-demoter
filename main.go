package main

import (
	"context"
	"flag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	match = flag.String("match", "", "Matches the user/group to the owner of the given path")
	debug = flag.Bool("debug", false, "Enable debug logging")
)

func main() {

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Requires command and its arguments to execute in demoted state")
	}

	uid, gid, err := resolveIds()
	if err != nil {
		log.WithError(err).Fatal("Failed to resolve IDs")
	}

	err = runCommand(uid, gid, args)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		} else {
			log.WithError(err).Fatal("Failed to run sub-command")
		}
	}
}

func resolveIds() (uint32, uint32, error) {
	var matchStatT *syscall.Stat_t
	var uid, gid uint32
	if *match != "" {
		log.Debugf("Using %s to match uid and gid", *match)

		matchInfo, err := os.Stat(*match)
		if err != nil {
			return 0, 0, errors.Wrap(err, "unable to inspect match path")
		}

		var ok bool
		if matchStatT, ok = matchInfo.Sys().(*syscall.Stat_t); !ok {
			return 0, 0, errors.Errorf("unsupported file info stat type: %v", matchInfo.Sys())
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
			return 0, errors.Wrapf(err, "Invalid %s", idPart)
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
			return 0, errors.Errorf("unknown id part: %v", idPart)
		}

		log.Debugf("Resolved %s=%d from match path", idPart, desired)
		return desired, nil
	}

	return 0, nil
}

func runCommand(uid uint32, gid uint32, commandAndArgs []string) error {
	command := exec.Command(commandAndArgs[0], commandAndArgs[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if uid != 0 || gid != 0 {
		command.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uid,
				Gid: gid,
			},
		}
	}

	err := command.Start()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSignalForwarding(ctx, command)

	err = command.Wait()
	if err != nil {
		return err
	}

	return nil
}

func setupSignalForwarding(ctx context.Context, cmd *exec.Cmd) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)

	go func() {
		for {
			select {
			case sig := <-signals:
				err := cmd.Process.Signal(sig)
				if err != nil {
					log.WithError(err).Error("Failed to signal sub-command")
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
