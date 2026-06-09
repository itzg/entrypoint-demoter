package entrypoint_demoter

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func RunCommand(uid uint32, gid uint32, stdinOnTerm string, stdinOnTermAnnounce string, stdinOnTermDelay time.Duration, commandAndArgs []string) error {
	return RunCommandWithListeners(uid, gid, stdinOnTerm, stdinOnTermAnnounce, stdinOnTermDelay, commandAndArgs)
}

func RunCommandWithListeners(uid uint32, gid uint32, stdinOnTerm string, stdinOnTermAnnounce string, stdinOnTermDelay time.Duration, commandAndArgs []string, listeners ...StdInOutListener) error {
	command := exec.Command(commandAndArgs[0], commandAndArgs[1:]...)
	command.Stderr = os.Stderr

	stdinPipe, err := command.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to get stdin pipe: %w", err)
	}
	//noinspection GoUnhandledErrorResult
	defer stdinPipe.Close()

	go RunStdinPumper(stdinPipe)
	for _, listener := range listeners {
		listener.UseStdin(stdinPipe)
	}

	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to get stdout pipe: %w", err)
	}

	if uid != 0 || gid != 0 {
		setCredentials(uid, gid, command)
	}

	err = command.Start()
	if err != nil {
		return err
	}

	go FanoutStdout(stdoutPipe, listeners...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalForwarding(ctx, command, stdinOnTerm, stdinOnTermAnnounce, stdinOnTermDelay, stdinPipe)

	err = command.Wait()
	if err != nil {
		return err
	}

	return nil
}

func setupSignalForwarding(ctx context.Context, cmd *exec.Cmd, stdinOnTerm string, stdinOnTermAnnounce string, stdinOnTermDelay time.Duration, stdinPipe io.Writer) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-signals:
				log.WithField("signal", sig).Debug("Forwarding signal")

				if stdinOnTerm != "" && sig == syscall.SIGTERM {
					if stdinOnTermAnnounce != "" {
						announce := strings.ReplaceAll(stdinOnTermAnnounce, "%delay%",
							strconv.FormatInt(int64(stdinOnTermDelay/time.Second), 10))
						log.WithField("message", announce).Debug("Sending announce on stdin due to SIGTERM")
						stdinPipe.Write([]byte(announce + "\n"))
					}

					if stdinOnTermDelay > 0 {
						// Wait before sending the stop message, but let a second TERM
						// (e.g. an impatient `docker stop`) short-circuit the wait so we
						// never block past the container's stop grace period.
						timer := time.NewTimer(stdinOnTermDelay)
						select {
						case <-timer.C:
						case <-signals:
							timer.Stop()
							log.Debug("Second TERM received; skipping remaining delay")
						case <-ctx.Done():
							timer.Stop()
							return
						}
					}

					log.WithField("message", stdinOnTerm).Debug("Sending message on stdin due to SIGTERM")
					stdinPipe.Write([]byte(stdinOnTerm + "\n"))
					continue
				}

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
