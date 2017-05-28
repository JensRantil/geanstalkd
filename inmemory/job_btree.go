package inmemory

import (
	"sync"

	"github.com/google/btree"

	"github.com/JensRantil/geanstalkd"
)

// jobById btree interface
type jobIDJobBTreeItem struct {
	item *geanstalkd.Job
}

func (a jobIDJobBTreeItem) Less(b btree.Item) bool {
	return a.item.ID < b.(jobIDJobBTreeItem).item.ID
}

// BTreeJobRegistry implements a JobRegistry backed by a BTree.
type BTreeJobRegistry struct {
	btree *btree.BTree
	lock  sync.RWMutex
}

// NewBTreeJobRegistry returns a new BTreeJobRegistry backed by btree.
func NewBTreeJobRegistry(btree *btree.BTree) *BTreeJobRegistry {
	return &BTreeJobRegistry{
		btree: btree,
	}
}

// Insert inserts a new job. It returns geanstalkd.ErrJobAlreadyExist if a job
// with the same ID already has been inserted.
func (i *BTreeJobRegistry) Insert(j *geanstalkd.Job) error {
	item := jobIDJobBTreeItem{j}

	i.lock.Lock()
	defer i.lock.Unlock()

	if i.btree.Has(item) {
		return geanstalkd.ErrJobAlreadyExist
	}

	i.btree.ReplaceOrInsert(item)
	return nil
}

// Update updates a previously inserted job. It returns
// geanstalkd.ErrJobMissing if it can't find a job with the given ID.
func (i *BTreeJobRegistry) Update(j *geanstalkd.Job) error {
	item := jobIDJobBTreeItem{j}

	i.lock.Lock()
	defer i.lock.Unlock()

	if !i.btree.Has(item) {
		return geanstalkd.ErrJobMissing
	}

	i.btree.ReplaceOrInsert(item)
	return nil
}

func itemToJob(item btree.Item) *geanstalkd.Job {
	return item.(jobIDJobBTreeItem).item
}

// GetByID queries a job with the given JobID. It returns the
// geanstalkd.ErrJobMissing error if the job could not be found.
func (i *BTreeJobRegistry) GetByID(id geanstalkd.JobID) (*geanstalkd.Job, error) {
	key := jobIDJobBTreeItem{&geanstalkd.Job{ID: id}}

	i.lock.RLock()
	item := i.btree.Get(key)
	i.lock.RUnlock()
	if item == nil {
		return nil, geanstalkd.ErrJobMissing
	}

	return itemToJob(item), nil
}

// DeleteByID deletes a job previously inserted. It returns
// geanstalkd.ErrJobMissing if the job could not be found.
func (i *BTreeJobRegistry) DeleteByID(id geanstalkd.JobID) error {
	key := jobIDJobBTreeItem{&geanstalkd.Job{ID: id}}

	i.lock.Lock()
	defer i.lock.Unlock()

	if item := i.btree.Delete(key); item == nil {
		return geanstalkd.ErrJobMissing
	}
	return nil
}

// GetLargestID returns the largest JobID for the jobs stored in this
// BTreeJobRegistry.
func (i *BTreeJobRegistry) GetLargestID() (geanstalkd.JobID, error) {
	i.lock.RLock()
	max := i.btree.Max()
	i.lock.RUnlock()

	if max == nil {
		return 0, geanstalkd.ErrEmptyRegistry
	}
	return itemToJob(max).ID, nil
}
