package main

import (
	"testing"
)

func TestBuildOptions(t *testing.T) {
	opts := map[string]interface{}{
		"landscape": true,
		"scale":     1.5,
	}

	if opts["landscape"] != true {
		t.Errorf("Expected landscape to be true")
	}

	if opts["scale"] != 1.5 {
		t.Errorf("Expected scale to be 1.5")
	}
}
