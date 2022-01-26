package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ConnString = ""

const WorkspaceEnvVar = "MSF_WORKSPACE"

func Connect(ctx context.Context) (*gorm.DB, int, error) {
	if ConnString != "" {
		log.Infof("Connecting with PostgreSQL connection string: %q", ConnString)
	}

	pgxCfg, err := pgx.ParseConfig(ConnString)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing PostgreSQL connection string %q: %w", ConnString, err)
	}
	log.Infof("Connected to Metasploit PostgreSQL database %q at %s:%d as user %q", pgxCfg.Database, pgxCfg.Host, pgxCfg.Port, pgxCfg.User)

	sqlConn := stdlib.OpenDB(*pgxCfg)
	gormDb, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlConn}), &gorm.Config{})
	if err != nil {
		return nil, 0, fmt.Errorf("connect to PostgreSQL database: %w", err)
	}

	// use this to print all queries
	// gormDb = gormDb.Debug()

	workspace := os.Getenv(WorkspaceEnvVar)
	if workspace == "" {
		workspace = "default"
	}

	workspaceId, err := GetWorkspaceId(gormDb, workspace)
	if err != nil {
		return nil, 0, fmt.Errorf("reading ID of Metasploit workspace %q: %w", workspace, err)
	}

	log.Debugf("ID of Metasploit workspace %q: %d", workspace, workspaceId)

	return gormDb, workspaceId, nil
}
