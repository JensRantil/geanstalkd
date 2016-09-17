package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
)

// TODO: All caps.
const maxLineLength = 1024

type tcpListener struct {
	server *server
}

func (tl *tcpListener) Serve(ctx context.Context, listenAddr string) {

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
	Server *server

	Ctx             context.Context
	CloseConnection context.CancelFunc

	Conn *textproto.Conn
}

func (ch connectionHandler) Handle() {
	defer ch.CloseConnection()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-ch.Ctx.Done()

		// This means things will fail inside handleSingleRequest() immediately.
		ch.Conn.Close()

		wg.Done()
	}()
	defer wg.Wait()

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

	var wg sync.WaitGroup
	defer wg.Wait()

	var commandLine string
	if commandLine, err = readCappedLine(ch.Conn.Reader.R, maxLineLength); err != nil {
		ch.CloseConnection()
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

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
			case "delete":
				handler = deleteHandler
			}
		}

		handler(ch, id, cmdArgs)
		ch.Conn.Pipeline.EndResponse(id)
	}()

}

func quitHandler(ch connectionHandler, pipelineId uint, cmdArgs cmdArgs) {
	ch.Conn.Pipeline.EndRequest(pipelineId)
	ch.Conn.Pipeline.StartResponse(pipelineId)
	ch.CloseConnection()
}

type integerParser struct {
	Err error
}

func (i *integerParser) Parse(s string) uint64 {
	result, err := strconv.ParseUint(s, 10, 64)
	if i.Err != nil {
		i.Err = err
	}
	return result
}

func fillBuffer(b []byte, source io.Reader) error {
	from := 0
	to := len(b)
	for from < to {
		nread, err := source.Read(b[from:to])
		if err != nil {
			return err
		}
		from += nread
	}
	return nil
}

func putHandler(ch connectionHandler, pipelineId uint, cmdArgs cmdArgs) {
	if len(cmdArgs) != 4 {
		ch.Conn.Pipeline.EndRequest(pipelineId)
		ch.Conn.Pipeline.StartResponse(pipelineId)
		ch.Conn.Writer.PrintfLine("BAD_FORMAT")
		return
	}

	p := new(integerParser)
	pri := p.Parse(cmdArgs[0])
	delay := p.Parse(cmdArgs[1])
	ttr := p.Parse(cmdArgs[2])
	nbytes := p.Parse(cmdArgs[3])
	if p.Err != nil {
		ch.Conn.Pipeline.EndRequest(pipelineId)
		ch.Conn.Pipeline.StartResponse(pipelineId)
		ch.Conn.Writer.PrintfLine("BAD_FORMAT")
		return
	}

	// Read up job data

	jobdata := make([]byte, nbytes)
	fillBuffer(jobdata, ch.Conn.Reader.R)
	if additionalData, err := readCappedLine(ch.Conn.Reader.R, maxLineLength); err != nil || len(additionalData) != 0 {
		// There was more data than expected.
		ch.Conn.Pipeline.EndRequest(pipelineId)
		ch.Conn.Pipeline.StartResponse(pipelineId)
		ch.Conn.Writer.PrintfLine("EXPECTED_CRLF")
		return
	}

	ch.Conn.Pipeline.EndRequest(pipelineId)

	job := ch.Server.BuildJob(
		priority(pri),
		time.Now().Add(time.Duration(delay)*time.Second),
		time.Duration(ttr)*time.Second,
		jobdata,
	)
	if err := ch.Server.Add(job); err != nil {
		if err == drainingError {
			ch.Conn.Pipeline.EndRequest(pipelineId)
			ch.Conn.Pipeline.StartResponse(pipelineId)
			ch.Conn.Writer.PrintfLine("DRAINING")
			return
		}
		log.Fatalln(err)
	}

	ch.Conn.Pipeline.StartResponse(pipelineId)
	ch.Conn.Writer.PrintfLine("INSERTED %d", job.Id)
}

func deleteHandler(ch connectionHandler, pipelineId uint, cmdArgs cmdArgs) {
	if len(cmdArgs) != 1 {
		ch.Conn.Pipeline.EndRequest(pipelineId)
		ch.Conn.Pipeline.StartResponse(pipelineId)
		ch.Conn.Writer.PrintfLine("BAD_FORMAT")
		return
	}

	p := new(integerParser)
	id := p.Parse(cmdArgs[0])
	if p.Err != nil {
		ch.Conn.Pipeline.EndRequest(pipelineId)
		ch.Conn.Pipeline.StartResponse(pipelineId)
		ch.Conn.Writer.PrintfLine("BAD_FORMAT")
		return
	}

	ch.Conn.Pipeline.EndRequest(pipelineId)

	err := ch.Server.Delete(jobId(id))

	ch.Conn.Pipeline.StartResponse(pipelineId)

	if err != nil {
		ch.Conn.Writer.PrintfLine("NOT_FOUND")
		return
	}

	ch.Conn.Writer.PrintfLine("DELETED")
}

func unknownCommandHandler(ch connectionHandler, pipelineId uint, cmdArgs cmdArgs) {
	ch.Conn.Pipeline.EndRequest(pipelineId)
	ch.Conn.Pipeline.StartResponse(pipelineId)
	ch.Conn.Writer.PrintfLine("UNKNOWN_COMMAND")
}
