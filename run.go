package entrypoint_demoter

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func RunCommand(uid uint32, gid uint32, stdinOnTerm string, commandAndArgs []string) error {
	return RunCommandWithListeners(uid, gid, stdinOnTerm, commandAndArgs)
}

func RunCommandWithListeners(uid uint32, gid uint32, stdinOnTerm string, commandAndArgs []string, listeners ...StdInOutListener) error {
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

	setupSignalForwarding(ctx, command, stdinOnTerm, stdinPipe)

	err = command.Wait()
	if err != nil {
		return err
	}

	return nil
}

func setupSignalForwarding(ctx context.Context, cmd *exec.Cmd, stdinOnTerm string, stdinPipe io.Writer) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-signals:
				log.WithField("signal", sig).Debug("Forwarding signal")

				if stdinOnTerm != "" && sig == syscall.SIGTERM {
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
