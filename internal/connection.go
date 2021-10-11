package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

var connString = ""

func Connect(ctx context.Context) (*pgx.Conn, int, error) {
	if connString != "" {
		log.Infof("Connecting with PostgreSQL connection string: %q", connString)
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, 0, fmt.Errorf("connecting to PostgreSQL database: %w", err)
	}

	cfg := conn.Config()
	log.Infof("Connected to Metasploit PostgreSQL database %q at %s:%d as %q", cfg.Database, cfg.Host, cfg.Port, cfg.User)

	workspace := os.Getenv("MSF_WORKSPACE")
	if workspace == "" {
		workspace = "default"
	}

	workspaceId, err := GetWorkspaceId(ctx, conn, workspace)
	if err != nil {
		return nil, 0, fmt.Errorf("reading ID of workspace %q: %w", workspace, err)
	}

	log.Debugf("ID of workspace %q: %d", workspace, workspaceId)

	return conn, workspaceId, nil
}
