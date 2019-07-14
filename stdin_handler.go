package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type StdinHandler struct {
	stdinMux    chan []byte
	noWarnStdin bool
}

func NewStdinHandler(noWarnStdin bool) *StdinHandler {
	return &StdinHandler{
		noWarnStdin: noWarnStdin,
		stdinMux:    make(chan []byte, 1),
	}
}

func (h *StdinHandler) Handle(ctx context.Context, stdinPipe io.Writer) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case chunk := <-h.stdinMux:
				_, err := stdinPipe.Write(chunk)
				if err != nil {
					log.WithError(err).Warn("stdin failed to write to pipe")
					return
				}
			}
		}
	}()

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if !h.noWarnStdin {
					log.WithError(err).Warn("failed to read stdin")
				}
				return
			}
			h.stdinMux <- buf[:n]
		}
	}()
}

func (h *StdinHandler) Send(data []byte) {
	h.stdinMux <- data
}
