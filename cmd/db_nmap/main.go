package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/jackc/pgx/v4"
	"github.com/jojonas/db_nmap/internal"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var connString = ""

func main() {
	log.SetLevel(logrus.DebugLevel)

	ctx := context.Background()

	if connString != "" {
		log.Infof("Connecting with PostgreSQL connection string: %q", connString)
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL database: %v", err)
	}
	defer conn.Close(ctx)

	cfg := conn.Config()
	log.Infof("Connected to Metasploit PostgreSQL database %q at %s:%d as %q", cfg.Database, cfg.Host, cfg.Port, cfg.User)

	workspace := os.Getenv("MSF_WORKSPACE")
	if workspace == "" {
		workspace = "default"
	}

	workspaceId, err := internal.GetWorkspaceId(ctx, conn, workspace)
	if err != nil {
		log.Fatalf("Unable to read ID of workspace %q: %v", workspace, err)
	}

	log.Debugf("ID of workspace %q: %d", workspace, workspaceId)

	cmd := exec.CommandContext(ctx, "/usr/bin/nmap", os.Args[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	hostCount := 0
	serviceCount := 0
	err = runNmap(cmd, func(host internal.NmapHost) error {
		n, err := internal.InsertHost(ctx, conn, int(workspaceId), host)

		if err != nil {
			log.Warnf("Inserting host into DB: %v", err)
		}

		if n > 0 {
			hostCount += 1
			serviceCount += n
		}

		return nil
	})

	log.Infof("Wrapper stats: registered %d hosts with %d services.", hostCount, serviceCount)

	if err != nil {
		log.Fatalf("%v", err)
	}
}
