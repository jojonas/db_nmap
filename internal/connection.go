package internal

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
)

var ConnString = ""

const WorkspaceEnvVar = "MSF_WORKSPACE"

func Connect(ctx context.Context) (*pgx.Conn, int, error) {
	if ConnString != "" {
		log.Infof("Connecting with PostgreSQL connection string: %q", ConnString)
	}

	conn, err := pgx.Connect(ctx, ConnString)
	if err != nil {
		return nil, 0, fmt.Errorf("connecting to PostgreSQL database: %w", err)
	}

	cfg := conn.Config()
	log.Infof("Connected to Metasploit PostgreSQL database %q at %s:%d as %q", cfg.Database, cfg.Host, cfg.Port, cfg.User)

	workspace := os.Getenv(WorkspaceEnvVar)
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

func CheckForReconnect(ctx context.Context, conn *pgx.Conn, retries int) (*pgx.Conn, error) {
	if conn.Ping(ctx) == nil {
		return conn, nil
	}

	log.Debug("Reconnecting...")

	var err error
	var newConn *pgx.Conn

	for retry := 0; retry < retries; retry++ {
		newConn, err = pgx.ConnectConfig(ctx, conn.Config())
		if err == nil {
			return newConn, nil
		}

		// sleep 1s, 2s, 4s, 8s, 16s, 32s, ...
		backoff := 1 * time.Second * time.Duration(math.Pow(2, float64(retry)))
		log.Debugf("Reconnect failed, waiting %.1fs for the next attempt...", backoff.Seconds())
		time.Sleep(backoff)
	}

	return nil, fmt.Errorf("connect: %w", err)
}
