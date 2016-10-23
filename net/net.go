// Implements all network communication.
package net

import (
	"context"
	"io"
	"log"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JensRantil/geanstalkd"
)

// TODO: All caps.
const maxLineLength = 1024

type Listener struct {
	Server *geanstalkd.Server
}

func (tl *Listener) Serve(ctx context.Context, l net.Listener) {
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
					tl.Server,
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
	Server *geanstalkd.Server

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
// ch.conn.Pipeline.EndRequest(pipelineID) and calling
// ch.conn.Pipeline.BeginResponse(pipelineID). Logic executed before
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

		var handler cmdHandler
		handler = unknownCommandHandler
		var cmdArgs []string

		if len(cmdAndArgs) > 0 {
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

func quitHandler(ch connectionHandler, pipelineID uint, cmdArgs cmdArgs) {
	ch.Conn.Pipeline.EndRequest(pipelineID)
	ch.Conn.Pipeline.StartResponse(pipelineID)
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

func putHandler(ch connectionHandler, pipelineID uint, cmdArgs cmdArgs) {
	if len(cmdArgs) != 4 {
		ch.Conn.Pipeline.EndRequest(pipelineID)
		ch.Conn.Pipeline.StartResponse(pipelineID)
		ch.Conn.Writer.PrintfLine("BAD_FORMAT")
		return
	}

	p := new(integerParser)
	pri := p.Parse(cmdArgs[0])
	delay := p.Parse(cmdArgs[1])
	ttr := p.Parse(cmdArgs[2])
	nbytes := p.Parse(cmdArgs[3])
	if p.Err != nil {
		ch.Conn.Pipeline.EndRequest(pipelineID)
		ch.Conn.Pipeline.StartResponse(pipelineID)
		ch.Conn.Writer.PrintfLine("BAD_FORMAT")
		return
	}

	// Read up job data

	jobdata := make([]byte, nbytes)
	io.ReadFull(ch.Conn.Reader.R, jobdata)
	if additionalData, err := readCappedLine(ch.Conn.Reader.R, maxLineLength); err != nil || len(additionalData) != 0 {
		// There was more data than expected.
		ch.Conn.Pipeline.EndRequest(pipelineID)
		ch.Conn.Pipeline.StartResponse(pipelineID)
		ch.Conn.Writer.PrintfLine("EXPECTED_CRLF")
		return
	}

	ch.Conn.Pipeline.EndRequest(pipelineID)

	job := ch.Server.BuildJob(
		geanstalkd.Priority(pri),
		time.Now().Add(time.Duration(delay)*time.Second),
		time.Duration(ttr)*time.Second,
		jobdata,
	)
	if err := ch.Server.Add(job); err != nil {
		if err == geanstalkd.ErrDraining {
			ch.Conn.Pipeline.EndRequest(pipelineID)
			ch.Conn.Pipeline.StartResponse(pipelineID)
			ch.Conn.Writer.PrintfLine("DRAINING")
			return
		}
		log.Fatalln(err)
	}

	ch.Conn.Pipeline.StartResponse(pipelineID)
	ch.Conn.Writer.PrintfLine("INSERTED %d", job.ID)
}

func deleteHandler(ch connectionHandler, pipelineID uint, cmdArgs cmdArgs) {
	if len(cmdArgs) != 1 {
		ch.Conn.Pipeline.EndRequest(pipelineID)
		ch.Conn.Pipeline.StartResponse(pipelineID)
		ch.Conn.Writer.PrintfLine("BAD_FORMAT")
		return
	}

	p := new(integerParser)
	id := p.Parse(cmdArgs[0])
	if p.Err != nil {
		ch.Conn.Pipeline.EndRequest(pipelineID)
		ch.Conn.Pipeline.StartResponse(pipelineID)
		ch.Conn.Writer.PrintfLine("BAD_FORMAT")
		return
	}

	ch.Conn.Pipeline.EndRequest(pipelineID)

	err := ch.Server.Delete(geanstalkd.JobID(id))

	ch.Conn.Pipeline.StartResponse(pipelineID)

	if err != nil {
		ch.Conn.Writer.PrintfLine("NOT_FOUND")
		return
	}

	ch.Conn.Writer.PrintfLine("DELETED")
}

func unknownCommandHandler(ch connectionHandler, pipelineID uint, cmdArgs cmdArgs) {
	ch.Conn.Pipeline.EndRequest(pipelineID)
	ch.Conn.Pipeline.StartResponse(pipelineID)
	ch.Conn.Writer.PrintfLine("UNKNOWN_COMMAND")
}
