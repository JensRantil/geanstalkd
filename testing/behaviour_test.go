package testing

import (
	. "testing"
)

// Testing the test :-) Had some issues debugging a test so added this just in
// case.
func TestTimes(t *T) {
	if !earlyTime.Before(laterTime) {
		t.Error("Order incorrect.")
	}
}
