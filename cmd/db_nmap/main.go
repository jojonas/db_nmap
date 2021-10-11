package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/jojonas/db_nmap/internal"
)

var log = internal.Logger
var binaryPath = "nmap"

func main() {
	ctx := context.Background()

	conn, workspaceId, err := internal.Connect(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close(ctx)

	cmd := exec.CommandContext(ctx, binaryPath, os.Args[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	hostCount := 0
	serviceCount := 0
	err = runNmap(cmd, func(host internal.NmapHost) error {
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

	log.Infof("Wrapper stats: registered %d hosts with %d services.", hostCount, serviceCount)

	if err != nil {
		log.Errorf("%v", err)
	}

	os.Exit(cmd.ProcessState.ExitCode())
}
