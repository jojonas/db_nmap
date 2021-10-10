package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	log.SetLevel(logrus.DebugLevel)

	workspace := "default"

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	cfg := conn.Config()
	log.Infof("Connected to Metasploit database at host=%s port=%d database=%s !", cfg.Host, cfg.Port, cfg.Database)

	row := conn.QueryRow(ctx, "select id from workspaces where name=$1", workspace)
	var workspaceId int
	err = row.Scan(&workspaceId)
	if err != nil {
		log.Fatalf("Unable to read ID of workspace %q: %v", workspace, err)
	}

	log.Debugf("ID of workspace %q: %d", workspace, workspaceId)

	cmd := exec.CommandContext(ctx, "/usr/bin/nmap", os.Args[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	hostCount := 0
	serviceCount := 0
	err = runNmap(cmd, func(host NmapHost) error {
		n, err := insertHost(ctx, conn, int(workspaceId), host)

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
