package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jojonas/db_nmap/internal"
)

const multipleHostnamesEnvVariable = "DB_NMAP_ENABLE_MULTIPLE_HOSTNAMES"

var (
	log                     = internal.Logger
	version                 = "dev"
	enableMultipleHostnames = ""
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s XMLFILE [XMLFILE...]\n", os.Args[0])
		os.Exit(1)
	}

	b := asBool(enableMultipleHostnames)
	if b != nil {
		internal.EnableMultipleHostnames = *b
	}

	v, ok := os.LookupEnv(multipleHostnamesEnvVariable)
	if ok {
		b := asBool(v)
		if b != nil {
			internal.EnableMultipleHostnames = *b
		}
	}

	log.Infof("db_import %s starting...", version)

	ctx := context.Background()

	db, workspaceId, err := internal.Connect(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	hostCount := 0
	serviceCount := 0

	for _, filename := range os.Args[1:] {
		file, err := os.Open(filename)
		if err != nil {
			log.Errorf("Opening %q: %v", filename, err)
			continue
		}

		err = internal.ParseNmapXML(file, func(host internal.NmapHost) error {
			n, err := internal.InsertHost(db, int(workspaceId), host)
			if err != nil {
				log.Warnf("Inserting host into DB: %v", err)
				return nil
			}

			if n > 0 {
				hostCount += 1
				serviceCount += n
			}

			return nil
		})

		if err != nil {
			log.Errorf("Parsing %q: %v", filename, err)
		}
	}

	log.Infof("Import stats: registered %d hosts with %d services.", hostCount, serviceCount)
}

func asBool(s string) *bool {
	t := true
	f := false

	switch strings.ToLower(s) {
	case "y", "yes", "true", "1":
		return &t
	case "n", "no", "false", "0":
		return &f
	default:
		return nil
	}
}
