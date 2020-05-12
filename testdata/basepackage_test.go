package testdata

import (
	"os"
	"testing"
)


func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestBasePackageFunc(t *testing.T) {
	if BasePackageFunc() != baseMessage {
		t.Fatal("Unexpected message was received.")
	}
}
