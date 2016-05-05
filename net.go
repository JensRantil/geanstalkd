package main

import (
	"fmt"
	"net"
	"net/textproto"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/context"
)

type tcpListener struct {
	server server
}

func (tl tcpListener) Serve(ctx context.Context, listenAddr string) {

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, listenAddr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				// This happens if the listener has been closed.
				return
			}
			// Handle connections in a new goroutine.
			go func() {
				wg.Add(1)
				defer wg.Done()

				childCtx, cancel := context.WithCancel(ctx)
				ch := connectionHandler{
					tl.server,
					childCtx,
					cancel,
					textproto.NewConn(conn),
				}
				ch.Handle()
			}()
		}
	}()

	<-ctx.Done()

	// Make sure we don't accept anymore connections
	l.Close()

	// Wait for all the connections have been closed
	wg.Wait()

}

type connectionHandler struct {
	Server server

	Ctx    context.Context
	Cancel context.CancelFunc

	Conn *textproto.Conn
}

func (ch connectionHandler) Handle() {
	defer ch.Cancel()

	go func() {
		<-ch.Ctx.Done()

		// This means things will fail inside handleSingleRequest() immediately.
		ch.Conn.Close()
	}()

	for {
		ch.handleSingleRequest()

		// Check if we should quit.
		select {
		case <-ch.Ctx.Done():
			return
		default:
		}
	}
}

type cmdArgs []string

// Responsible for making optional additional reads, calling
// ch.conn.Pipeline.EndRequest(pipelineId) and calling
// ch.conn.Pipeline.BeginResponse(pipelineId). Logic executed before
// BeginResponse will be executed in parallel. Logic executed after
// BeginResponse will be executed serially (useful for statistics for example).
type cmdHandler func(connectionHandler, uint, cmdArgs)

func (ch connectionHandler) handleSingleRequest() {
	var err error

	id := ch.Conn.Pipeline.Next()
	ch.Conn.Pipeline.StartRequest(id)

	go func() {
		var commandLine string
		if commandLine, err = ch.Conn.Reader.ReadLine(); err != nil {

			// Must be called to avoid deadlock in handleSingleRequest() when it calls ch.Conn.Pipeline.StartRequest(...)
			ch.Conn.Pipeline.EndRequest(id)
			ch.Conn.Pipeline.StartResponse(id)
			ch.Conn.Pipeline.EndResponse(id)

			return
		}

		cmdAndArgs := strings.Split(commandLine, " ")

		handler := unknownCommandHandler
		cmdArgs := make([]string, 0)

		if len(cmdAndArgs) != 0 {
			cmd := cmdAndArgs[0]
			cmdArgs = cmdAndArgs[1:]
			switch cmd {
			case "quit":
				handler = quitHandler
			case "put":
				handler = putHandler
			}
		}

		handler(ch, id, cmdArgs)
		ch.Conn.Pipeline.EndResponse(id)
	}()

}

func quitHandler(ch connectionHandler, pipelineId uint, cmdArgs cmdArgs) {
	ch.Conn.Pipeline.EndRequest(pipelineId)
	ch.Conn.Pipeline.StartResponse(pipelineId)
	ch.Cancel()
}

func putHandler(ch connectionHandler, pipelineId uint, cmdArgs cmdArgs) {
	ch.Conn.Pipeline.EndRequest(pipelineId)
	ch.Conn.Pipeline.StartResponse(pipelineId)
	ch.Conn.Writer.PrintfLine("PUT")
}

func unknownCommandHandler(ch connectionHandler, pipelineId uint, cmdArgs cmdArgs) {
	ch.Conn.Pipeline.EndRequest(pipelineId)
	ch.Conn.Pipeline.StartResponse(pipelineId)
	ch.Conn.Writer.PrintfLine("UNKNOWN_COMMAND")
}
