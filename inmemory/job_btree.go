package inmemory

import (
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

// BtreeJobRegistry implements a JobRegistry backed by a BTree.
type BTreeJobRegistry btree.BTree

func registryToBTree(r *BTreeJobRegistry) *btree.BTree {
	return (*btree.BTree)(r)
}

func (i *BTreeJobRegistry) Insert(j *geanstalkd.Job) error {
	item := jobIDJobBTreeItem{j}

	if registryToBTree(i).Has(item) {
		return geanstalkd.ErrJobAlreadyExist
	}

	registryToBTree(i).ReplaceOrInsert(item)
	return nil
}

func (i *BTreeJobRegistry) Update(j *geanstalkd.Job) error {
	item := jobIDJobBTreeItem{j}

	if !registryToBTree(i).Has(item) {
		return geanstalkd.ErrJobMissing
	}

	registryToBTree(i).ReplaceOrInsert(item)
	return nil
}

func itemToJob(item btree.Item) *geanstalkd.Job {
	return item.(jobIDJobBTreeItem).item
}

func (i *BTreeJobRegistry) GetByID(id geanstalkd.JobID) (*geanstalkd.Job, error) {
	key := jobIDJobBTreeItem{&geanstalkd.Job{ID: id}}
	item := registryToBTree(i).Get(key)
	if item == nil {
		return nil, geanstalkd.ErrJobMissing
	}

	return itemToJob(item), nil
}

func (i *BTreeJobRegistry) DeleteByID(id geanstalkd.JobID) error {
	key := jobIDJobBTreeItem{&geanstalkd.Job{ID: id}}
	if item := registryToBTree(i).Delete(key); item == nil {
		return geanstalkd.ErrJobMissing
	}
	return nil
}

func (i *BTreeJobRegistry) GetLargestID() (geanstalkd.JobID, error) {
	max := registryToBTree(i).Max()
	if max == nil {
		return 0, geanstalkd.ErrEmptyRegistry
	}
	return itemToJob(max).ID, nil
}
