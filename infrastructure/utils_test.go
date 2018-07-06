package infrastructure

import (
	"testing"
)

func TestMainFunc(t *testing.T) {
	tls := developmentTLSConfig()
	if tls == nil {
		t.Errorf("no error config generated")
	}

}
