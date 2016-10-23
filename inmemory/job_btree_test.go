package inmemory

import (
	. "testing"

	"github.com/JensRantil/geanstalkd"
)

func TestInMemoryBTreeJobRegistryImplementsJobRegistry(t *T) {
	t.Parallel()
	var _ geanstalkd.JobRegistry = (*BTreeJobRegistry)(nil)
}
