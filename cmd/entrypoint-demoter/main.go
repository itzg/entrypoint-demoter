package main

import (
	"flag"
	"fmt"
	"github.com/itzg/entrypoint-demoter"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
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
	stdinOnTerm = flag.String("stdin-on-term", "",
		"If set, the given content will be written to the sub-command's stdin when TERM signal is received")
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

	uid, gid, err := entrypoint_demoter.ResolveIds(*match)
	if err != nil {
		log.WithError(err).Fatal("Failed to resolve IDs")
	}

	err = entrypoint_demoter.RunCommand(uid, gid, *stdinOnTerm, args)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		} else {
			log.WithError(err).Fatal("Failed to run sub-command")
		}
	}
}
