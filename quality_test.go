package geanstalkd

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	. "testing"
)

func TestGoVet(t *T) {
	cmd := exec.Command("go", "vet", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Error("Something is broken. See other errors:\n" + string(output))
	}
}

var (
	golintWarning = regexp.MustCompile("^(?P<file>.*):(?P<line>\\d+):(?P<col>\\d+): (?P<desc>.*)$")

	// AFAIK golint doesn't support a way to ignore warnings. Doing it this way instead.
	golintIgnores = regexp.MustCompile(strings.Join([]string{
		"^testing/.*\\.go:\\d+:2: should not use dot imports$",
	}, "|"))
)

func TestGoLint(t *T) {
	cmd := exec.Command("golint", "-set_exit_status", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		buf := bytes.NewBuffer(output)
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			if match := golintWarning.FindString(scanner.Text()); match != "" {
				if ignoreCheck := golintIgnores.FindString(scanner.Text()); ignoreCheck == "" {
					t.Error(scanner.Text())
				}
			}
		}
	}
}
