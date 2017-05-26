package geanstalkd

import (
	"os/exec"
	. "testing"
)

func TestGoVet(t *T) {
	cmd := exec.Command("go", "vet", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Error("Something is broken. See other errors:\n" + string(output))
	}
}
