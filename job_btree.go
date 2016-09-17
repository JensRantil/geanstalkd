package main

import (
	"github.com/google/btree"
)

// jobById btree interface
type jobIdJobBTreeItem job

func (a jobIdJobBTreeItem) Less(b btree.Item) bool {
	return a.Id < b.(jobIdJobBTreeItem).Id
}
