package main

import (
	"os"
	"os/signal"

	"golang.org/x/net/context"
)

const (
	// TODO: Make this command line flag.
	ConnHost = "localhost"

	// TODO: Make this command line flag.
	ConnPort = "11300"

	ConnType = "tcp"
)

func cancelOnInterrupt(ctx context.Context, cancel func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()
}

func generateIds(ctx context.Context) <-chan jobID {
	ids := make(chan jobID, 100)
	go func() {
		nextID := jobID(1)
		for {
			select {
			case ids <- nextID:
				nextID++
			case <-ctx.Done():
				return
			}
		}
	}()

	return ids
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancelOnInterrupt(ctx, cancel)

	ids := generateIds(ctx)
	srv := newServer(ids)
	tcpListener := tcpListener{srv}

	tcpListener.Serve(ctx, ConnHost+":"+ConnPort)
}
