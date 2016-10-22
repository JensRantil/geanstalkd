package inmemory

import (
	. "testing"

	"github.com/JensRantil/geanstalkd"
)

func TestInMemoryBTreeJobRegistryImplementsJobRegistry(t T) {
	var _ geanstalkd.JobRegistry = InMemoryBTreeJobRegistry{nil}
}
