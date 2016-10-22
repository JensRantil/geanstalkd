package main

import (
	gonet "net"
	"os"
	"os/signal"

	"golang.org/x/net/context"

	"github.com/JensRantil/geanstalkd"
	"github.com/JensRantil/geanstalkd/inmemory"
	"github.com/JensRantil/geanstalkd/net"
	"github.com/google/btree"
)

const (
	// ConnHost is the host that the socket will listen on when the server is started.
	//
	// TODO: Make this command line flag.
	ConnHost = "localhost"

	// ConnPort is the port that the socket will listen on when the server is started.
	//
	// TODO: Make this command line flag.
	ConnPort = "11300"
)

func cancelOnInterrupt(ctx context.Context, cancel func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()
}

func generateIds(ctx context.Context) <-chan geanstalkd.JobID {
	ids := make(chan geanstalkd.JobID, 100)
	go func() {
		nextID := geanstalkd.JobID(1)
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

// DefaultBTreeDegree is the maximum number of items a BTree node holds.
const DefaultBTreeDegree = 16

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancelOnInterrupt(ctx, cancel)

	ids := generateIds(ctx)
	srv := &geanstalkd.Server{
		Storage: geanstalkd.NewLockService(
			&geanstalkd.StorageService{
				(*inmemory.BTreeJobRegistry)(btree.New(DefaultBTreeDegree)),
				inmemory.NewJobHeapPriorityQueue(),
				inmemory.NewJobHeapPriorityQueue(),
			},
		),
		Ids: ids,
	}
	connListener := net.Listener{srv}

	l, err := gonet.Listen("tcp", ConnHost+":"+ConnPort)
	if err != nil {
		panic(err)
	}

	connListener.Serve(ctx, l)
}
