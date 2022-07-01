package main

import (
	"testing"
)

func SampleTest(t *testing.T) {
	if 5 == 6 {
		t.Errorf("5 is not %d", 6)
	}
}
