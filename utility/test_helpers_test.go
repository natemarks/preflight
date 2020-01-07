package utility

import (
	"fmt"
	"testing"
)

func TestThisFunctionName(t *testing.T) {
	got := ThisFunctionName()
	if got != "utility.TestThisFunctionName" {
		t.Errorf("ThisFunctionName() = %s; want utility.TestThisFunctionName", got)
	}
}

func testCaller() string {
	return CallerFunctionName()
}

func TestCallerFunctionName(t *testing.T) {
	got := testCaller()
	if got != "utility.TestCallerFunctionName" {
		t.Errorf("CallerFunctionName() = %s; want utility.TestCallerFunctionName", got)
	}
}

func PrintHello() {
	fmt.Printf("my_stdout")
}
func TestCapOut(t *testing.T) {
	got := CapOut(PrintHello)
	if got != "my_stdout" {
		t.Fail()
	}
}
