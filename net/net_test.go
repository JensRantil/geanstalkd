package net

import (
	"bytes"
	"context"
	"net/textproto"

	"github.com/JensRantil/geanstalkd"
	"github.com/JensRantil/geanstalkd/inmemory"
	"github.com/google/btree"

	. "testing"
)

func TestJobIdIsIncreased(t *T) {
	t.Parallel()
	testInput("put 0 0 10 5\r\nhello\r\nput 0 0 10 5\r\nhello\r\n").ExpectingOutput(t, "INSERTED 1\r\nINSERTED 2\r\n")
}

func TestPut(t *T) {
	t.Parallel()
	testInput("put 0 0 10 5\r\nhello\r\n").ExpectingOutput(t, "INSERTED 1\r\n")
}

func TestPutWithBadFormat(t *T) {
	t.Parallel()
	testInput("put 0 0 10\r\n").ExpectingOutput(t, "BAD_FORMAT\r\n")
	testInput("put 0 0\r\n").ExpectingOutput(t, "BAD_FORMAT\r\n")
	testInput("put 0\r\n").ExpectingOutput(t, "BAD_FORMAT\r\n")
	testInput("put\r\n").ExpectingOutput(t, "BAD_FORMAT\r\n")
}

func TestUnknownCommand(t *T) {
	t.Parallel()
	testInput("this is a test\r\n").ExpectingOutput(t, "UNKNOWN_COMMAND\r\n")
}

func TestQuitCommand(t *T) {
	t.Parallel()
	testInput("quit\r\n").ExpectingOutput(t, "")
	testInput("quit\r\nthis is a test").ExpectingOutput(t, "")
}

type mockedReadWriteCloser struct {
	Input  *bytes.Buffer
	Closed bool
	output bytes.Buffer
}

func (m *mockedReadWriteCloser) Close() error {
	m.Closed = true
	return nil
}

func (m *mockedReadWriteCloser) Read(b []byte) (int, error) {
	return m.Input.Read(b)
}

func (m *mockedReadWriteCloser) Write(b []byte) (int, error) {
	return m.output.Write(b)
}

type inputOutputTest struct {
	mrwc *mockedReadWriteCloser
}

func testInput(input string) inputOutputTest {
	m := mockedReadWriteCloser{
		bytes.NewBufferString(input),
		false,
		bytes.Buffer{},
	}
	return inputOutputTest{&m}
}

const DefaultBTreeDegree = 16

func (iot inputOutputTest) ExpectingOutput(t *T, expected string) {
	ctx, cancel := context.WithCancel(context.Background())
	ids := geanstalkd.GenerateIds(ctx)
	srv := &geanstalkd.Server{
		Storage: geanstalkd.NewLockService(
			&geanstalkd.StorageService{
				Jobs:       inmemory.NewBTreeJobRegistry(btree.New(DefaultBTreeDegree)),
				ReadyQueue: inmemory.NewJobHeapPriorityQueue(),
				DelayQueue: inmemory.NewJobHeapPriorityQueue(),
			},
		),
		Ids: ids,
	}

	ch := connectionHandler{
		srv,
		ctx,
		cancel,
		textproto.NewConn(iot.mrwc),
	}
	ch.Handle()

	if !iot.mrwc.Closed {
		t.Error("Connection was not closed.")
	}

	if output := iot.mrwc.output.String(); output != expected {
		t.Errorf("Unexpected output. Output: %s Expected: %s", output, expected)
	}

	select {
	case <-ctx.Done():
	default:
		t.Error("Context wasn't done.")
	}
}
