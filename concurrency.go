package geanstalkd

import (
	"context"
	"sync"
)

// LockService handles long-polling and locking to orchestrate
// `StorageService`.
type LockService struct {
	storage *StorageService
	lock    sync.RWMutex
	cond    *channelCond
}

// NewLockService creates a new `LockService` which delegates the actual
// storage to storage.
func NewLockService(storage *StorageService) *LockService {
	ls := &LockService{
		storage: storage,
	}
	ls.cond = newChannelCond(&ls.lock)
	return ls
}

// Add adds a new job and notifies other goroutines that there is a new job
// available. If the storage returns an error, it is returned here.
func (ls *LockService) Add(j *Job) error {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	err := ls.storage.Add(j)
	if err == nil {
		// TODO: Target more specific goroutines.
		ls.cond.Broadcast()
	}

	return err
}

// Poll polls a new job. If there is no job available it waits for one to
// become available, or until the ctx is Done. Error is either an error
// returned from the storage's PopNextReady() call, or an error returns from
// context.
func (ls *LockService) Poll(ctx context.Context) (*Job, error) {
	ls.lock.Lock()

	for {
		job, err := ls.storage.PopNextReady()
		if err == nil {
			return job, nil
		} else if err != ErrNoJobReady {
			ls.lock.Unlock()
			return nil, err
		}

		ok := ls.cond.Wait(ctx)
		if timeout := !ok; timeout {
			ls.lock.Unlock()
			return nil, ctx.Err()
		}
	}
}

// Delete deletes a job with the given ID. If an error is returned, it has been
// relayed from the storage.Delete() call.
func (ls *LockService) Delete(id JobID) error {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	return ls.storage.Delete(id)
}

// channelCond is very similar to `sync.Cond`, but supports timeouts while waiting.
type channelCond struct {
	outerLock sync.Locker

	lock     sync.RWMutex
	channels []chan struct{}
}

func newChannelCond(l sync.Locker) *channelCond {
	return &channelCond{
		outerLock: l,
	}
}

type channelCondID int

func (cc *channelCond) register() channelCondID {
	cc.channels = append(cc.channels, make(chan struct{}, 1))
	return channelCondID(len(cc.channels) - 1)
}

func (cc *channelCond) unregister(id channelCondID) {
	c := cc.channels[id]
	close(c)

	// Delete index `id`. See https://github.com/golang/go/wiki/SliceTricks.
	cc.channels = append(cc.channels[:id], cc.channels[id+1:]...)
}

// Wait waits for a conditional or timeout to happen. If a conditional event is
// triggered, it returns `true`. Otherwise it returns `false`.
func (cc *channelCond) Wait(ctx context.Context) bool {
	cc.lock.Lock()
	id := cc.register()

	// Channel could move around in the array while we are `Wait`ing, so
	// important to extract it before running `select` below.
	c := cc.channels[id]

	cc.lock.Unlock()
	defer func() {
		cc.lock.Lock()
		cc.unregister(id)
		cc.lock.Unlock()
	}()

	cc.outerLock.Unlock()

	select {
	case _, ok := <-c:
		if ok {
			cc.outerLock.Lock()
		}
		return ok
	case <-ctx.Done():
		return false
	}

}

// Can easily implement a Signal method here. If written, not sure it needs to
// have the `default` case below, though.

// Tell all waiting go routines that a conditional event happened.
func (cc *channelCond) Broadcast() {
	cc.lock.RLock()
	for _, c := range cc.channels {
		select {
		case c <- struct{}{}:
		default:
		}
	}
	cc.lock.RUnlock()
}
