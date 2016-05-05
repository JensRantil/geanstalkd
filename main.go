package main

import (
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/net/context"
)

const (
	// TODO: Make this command line flag.
	CONN_HOST = "localhost"

	// TODO: Make this command line flag.
	CONN_PORT = "11300"

	CONN_TYPE = "tcp"
)

func main() {
	srv := server{}
	tcpListener := tcpListener{srv}

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()

	tcpListener.Serve(ctx, CONN_HOST+":"+CONN_PORT)

}
