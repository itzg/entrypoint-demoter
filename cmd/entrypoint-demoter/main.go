package main

import (
	"flag"
	"fmt"
	"github.com/itzg/entrypoint-demoter"
	"github.com/itzg/go-flagsfiller"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

var (
	version = ""
	commit  = ""
	date    = ""
	builtBy = ""
)

var config struct {
	Match       string `usage:"Matches the user/group to the owner of the given path"`
	Debug       bool   `usage:"Enable debug logging"`
	Version     bool   `usage:"Show version info and exit"`
	StdinOnTerm string `usage:"If set, the given content will be written to the sub-command's stdin when TERM signal is received"`

	StdinOnTermAnnounce string `usage:"If set, written to the sub-command's stdin when TERM is received, before any stdin-on-term-delay and the stdin-on-term content. A %delay% token is replaced with the whole-second value of stdin-on-term-delay."`

	StdinOnTermDelay time.Duration `usage:"If set, waited after TERM (and after writing any stdin-on-term-announce) before the stdin-on-term content is written. Keep shorter than the container's stop grace period."`
}

func main() {

	err := flagsfiller.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	if config.Version {
		fmt.Printf("Version=%s, commit=%s, date=%s, builtBy=%s\n",
			version, commit, date, builtBy)
		return
	}

	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Requires command and its arguments to execute in demoted state")
	}

	uid, gid, err := entrypoint_demoter.ResolveIds(config.Match)
	if err != nil {
		log.WithError(err).Fatal("Failed to resolve IDs")
	}

	err = entrypoint_demoter.RunCommand(uid, gid, config.StdinOnTerm, config.StdinOnTermAnnounce, config.StdinOnTermDelay, args)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		} else {
			log.WithError(err).Fatal("Failed to run sub-command")
		}
	}
}
