package samplePackage

import "testing"

func TestIdentity(t *testing.T) {
	if Identity() != "samplePackage" {
		t.Fatal("Identity failed!")
	}
}
