package testdata

import "testing"

func TestBasePackageFunc(t *testing.T) {
	if BasePackageFunc() != baseMessage {
		t.Fatal("Unexpected message was received.")
	}
}