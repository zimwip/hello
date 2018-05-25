package router

import (
	"testing"
)

func TestMainFunc(t *testing.T) {
	tls := mainTest()
	if tls == nil {
		t.Errorf("no error config generated")
	}

}
