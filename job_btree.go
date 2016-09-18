package main

import (
	"github.com/google/btree"
)

// jobById btree interface
type jobIDJobBTreeItem job

func (a jobIDJobBTreeItem) Less(b btree.Item) bool {
	return a.ID < b.(jobIDJobBTreeItem).ID
}
