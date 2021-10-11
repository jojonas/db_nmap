package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jojonas/db_nmap/internal"
)

var log = internal.Logger

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s XMLFILE [XMLFILE...]\n", os.Args[0])
		os.Exit(1)
	}

	ctx := context.Background()

	conn, workspaceId, err := internal.Connect(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close(ctx)

	hostCount := 0
	serviceCount := 0

	for _, filename := range os.Args[1:] {
		file, err := os.Open(filename)
		if err != nil {
			log.Errorf("Opening %q: %v", filename, err)
			continue
		}

		err = internal.ParseNmapXML(file, func(host internal.NmapHost) error {
			n, err := internal.InsertHost(ctx, conn, int(workspaceId), host)

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
