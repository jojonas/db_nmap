package internal

import (
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestParse(t *testing.T) {
	log.SetLevel(logrus.DebugLevel)

	files := []string{"testdata/scanme.xml", "testdata/localhost.xml"}

	for _, filename := range files {
		t.Run(filename, func(t *testing.T) {
			reader, err := os.Open(filename)
			if err != nil {
				t.Fatalf("Error opening %q: %v", filename, err)
			}
			defer reader.Close()

			hosts := 0
			err = ParseNmapXML(reader, func(host NmapHost) error {
				hosts++
				return nil
			})

			if err != nil {
				t.Errorf("Error parsing %q: %v", filename, err)
			}

			if hosts == 0 {
				t.Error("No hosts found")
			}
		})
	}
}

func TestCheckVersion(t *testing.T) {
	var hook *test.Hook
	log, hook = test.NewNullLogger()

	version := "1.23"
	checkVersion(version)
	if !strings.Contains(hook.LastEntry().Message, "not tested") {
		t.Errorf("version %s does not trigger a log warning", version)
	}
}
