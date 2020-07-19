package entrypoint_demoter

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type StdInOutListener interface {
	UseStdin(wr io.Writer)
	HandleStdout(line string)
}

func FanoutStdout(stdoutPipe io.Reader, listeners ...StdInOutListener) {
	scanner := bufio.NewScanner(stdoutPipe)
	for scanner.Scan() {
		line := scanner.Text()
		for _, listener := range listeners {
			listener.HandleStdout(line)
		}
		// ...and output to our stdout
		fmt.Println(line)
	}
}

func RunStdinPumper(wr io.Writer) {
	// setup a long-running copy of our stdin to the child's stdin...unless ours is closed
	_, err := io.Copy(wr, os.Stdin)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// container's don't usually have input opened, so just a debug if this fails
			log.Debug("stdin is detached, so forwarding is disabled")
		} else {
			log.WithError(err).Warn("failed to read stdin")
		}

		return
	}
}
