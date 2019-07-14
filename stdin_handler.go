package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type StdinHandler struct {
	stdinMux chan []byte
}

func NewStdinHandler() *StdinHandler {
	return &StdinHandler{
		stdinMux: make(chan []byte, 1),
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
				// container's don't usually have input opened, so just a debug if this fails
				log.WithError(err).Debug("failed to read stdin")
				return
			}
			h.stdinMux <- buf[:n]
		}
	}()
}

func (h *StdinHandler) Send(data []byte) {
	h.stdinMux <- data
}
