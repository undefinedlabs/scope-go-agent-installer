package instrumentedPackage

import (
	"os"
	"testing"
	
	"go.undefinedlabs.com/scopeagent"
)

func TestMain(m *testing.M) {
	os.Exit(scopeagent.Run(m))
}

func TestIdentity(t *testing.T) {
	if Identity() != "instrumentedPackage" {
		t.Fatal("Identity failed!")
	}
}
