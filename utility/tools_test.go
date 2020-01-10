package utility

import "testing"

func TestContains(t *testing.T) {
	ll := []string{"aaa", "bbb", "ccc"}
	if !Contains(ll, "aaa") {
		t.Fail()
	}
	if Contains(ll, "ddd") {
		t.Fail()
	}
}
