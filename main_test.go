package main

import (
	"strings"
	"testing"

	"github.com/natemarks/preflight/utility"
)

func TestRealMain(t *testing.T) {
	got := utility.CapOut(main)
	if !strings.Contains(got, " level=warning msg=\"No config file found\"") {
		t.Fail()
	}
}
