package main

import (
	"io/ioutil"
	"os"
	"testing"
)

const expectedBaseInstFile = `package testdata

import (
	_ "go.undefinedlabs.com/scopeagent/autoinstrument"
)
`

const expectedSampleInstFile = `package samplePackage

import (
	_ "go.undefinedlabs.com/scopeagent/autoinstrument"
)
`

func TestInstallerProcessor(t *testing.T) {
	baseInstFilePath := "./testdata/scope_pkg_testdata_test.go"
	sampleInstFilePath := "./testdata/samplePackage/scope_pkg_samplePackage_test.go"

	defer func() {
		// Remove test files
		_ = os.Remove(baseInstFilePath)
		_ = os.Remove(sampleInstFilePath)
	}()

	// Remove previous files if any
	_ = os.Remove(baseInstFilePath)
	_ = os.Remove(sampleInstFilePath)

	// Process test data
	processFolder("./testdata/")

	// Base instrumentation file
	baseInstFile, err := os.Open(baseInstFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer baseInstFile.Close()
	data, err := ioutil.ReadAll(baseInstFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expectedBaseInstFile {
		t.Fatal("the base package instrumentation file is different than expected")
	}

	// Sample instrumentation file
	sampleInstFile, err := os.Open(sampleInstFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer sampleInstFile.Close()
	data, err = ioutil.ReadAll(sampleInstFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expectedSampleInstFile {
		t.Fatal("the sample package instrumentation file is different than expected")
	}

	// Instrumented test package shouldn't exist
	_, err = os.Stat("./testdata/instrumentedPackage/scope_pkg_instrumentedPackage_test.go")
	if err == nil {
		t.Fatal("the instrumented package instrumentation file shouldn't exist")
	}
}
