package inmemory

import (
	. "testing"

	"github.com/google/btree"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/JensRantil/geanstalkd/testing"
)

func TestBTreeJobRegistry(t *T) {
	t.Parallel()

	Convey("Given a fresh BTreeJobRegistry", t, func() {
		bt := (*BTreeJobRegistry)(btree.New(btree.DefaultFreeListSize))
		testing.GenericJobRegistryTest(bt)
	})
}
