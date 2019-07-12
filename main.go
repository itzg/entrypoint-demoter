package main

import (
	"context"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/signal"
)

var (
	version = ""
	commit  = ""
	date    = ""
	builtBy = ""
)

var (
	match       = flag.String("match", "", "Matches the user/group to the owner of the given path")
	debug       = flag.Bool("debug", false, "Enable debug logging")
	showVersion = flag.Bool("version", false, "Show version info and exit")
)

func main() {

	flag.Parse()

	if *showVersion {
		fmt.Printf("Version=%s, commit=%s, date=%s, builtBy=%s\n",
			version, commit, date, builtBy)
		return
	}

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

func runCommand(uid uint32, gid uint32, commandAndArgs []string) error {
	command := exec.Command(commandAndArgs[0], commandAndArgs[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if uid != 0 || gid != 0 {
		setCredentials(uid, gid, command)
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
