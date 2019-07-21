package entrypoint_demoter

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func RunCommand(uid uint32, gid uint32, stdinOnTerm string, commandAndArgs []string) error {
	command := exec.Command(commandAndArgs[0], commandAndArgs[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	stdinPipe, err := command.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "unable to get stdin pipe")
	}
	//noinspection GoUnhandledErrorResult
	defer stdinPipe.Close()

	if uid != 0 || gid != 0 {
		setCredentials(uid, gid, command)
	}

	err = command.Start()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stdinHandler := NewStdinHandler()
	stdinHandler.Handle(ctx, stdinPipe)

	setupSignalForwarding(ctx, command, stdinOnTerm, stdinHandler)

	err = command.Wait()
	if err != nil {
		return err
	}

	return nil
}

func setupSignalForwarding(ctx context.Context, cmd *exec.Cmd, stdinOnTerm string, stdinHandler *StdinHandler) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)

	go func() {
		for {
			select {
			case sig := <-signals:
				log.WithField("signal", sig).Debug("Forwarding signal")

				if stdinOnTerm != "" && sig == syscall.SIGTERM {
					log.WithField("message", stdinOnTerm).Debug("Sending message on stdin due to SIGTERM")
					stdinHandler.Send([]byte(stdinOnTerm + "\n"))
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
