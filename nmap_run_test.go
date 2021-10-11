package main

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestParse(t *testing.T) {
	log.SetLevel(logrus.DebugLevel)

	filename := "testdata/scanme.xml"
	reader, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Error opening %q: %v", filename, err)
	}
	defer reader.Close()

	hosts := 0
	err = parseNmapXML(reader, func(host NmapHost) error {
		hosts++
		return nil
	})

	if err != nil {
		t.Errorf("Error parsing %q: %v", filename, err)
	}

	if hosts == 0 {
		t.Error("No hosts fond.")
	}
}
