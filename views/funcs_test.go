package views

import (
	"strings"
	"testing"
)

func TestHumanizeBytes(t *testing.T) {

	e := humanizeBytes(42 * 1 << 20)
	if strings.Compare(e, "42MB") != 0 {
		t.Errorf("Wrong encode for MB %s\n", e)
	}

	e = humanizeBytes(3)
	if strings.Compare(e, "3B") != 0 {
		t.Error("Wrong encode for byte %s\n", e)
	}

	e = humanizeBytes(1<<62 + 1<<61)
	if strings.Compare(e, "6EB") != 0 {
		t.Error("Wrong max size %s\n", e)
	}
}
