package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ConnString = ""
var MetasploitDatabaseConfigurationFile = "/usr/share/metasploit-framework/config/database.yml"

const WorkspaceEnvVar = "MSF_WORKSPACE"

type MetasploitDatabaseConfigSection struct {
	Adapter  string
	Database string
	Username string
	Password string
	Host     string
	Port     uint16
	Pool     int
	Timeout  int
}

type MetasploitDatabaseConfig struct {
	Development MetasploitDatabaseConfigSection
	Production  MetasploitDatabaseConfigSection
	Test        MetasploitDatabaseConfigSection
}

func Connect(ctx context.Context) (*gorm.DB, int, error) {
	var err error
	var pgxCfg *pgx.ConnConfig

	if ConnString != "" {
		log.Infof("Connecting with PostgreSQL connection string: %q", ConnString)

		pgxCfg, err = pgx.ParseConfig(ConnString)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing PostgreSQL connection string %q: %w", ConnString, err)
		}
	} else {
		pgxCfg, err = readMetasploitConfiguration(MetasploitDatabaseConfigurationFile)

		if err != nil {
			if os.IsNotExist(err) {
				log.Info("Creating default config...")
				pgxCfg, err = pgx.ParseConfig("")
				if err != nil {
					return nil, 0, fmt.Errorf("creating default config: %w", err)
				}
			} else {
				return nil, 0, fmt.Errorf("parsing Metasploit database configuration file %q: %w", MetasploitDatabaseConfigurationFile, err)
			}
		} else {
			log.Infof("Read Metasploit database configuration file %q.", MetasploitDatabaseConfigurationFile)
		}
	}

	log.Infof("Connecting to Metasploit PostgreSQL database %q at %s:%d as user %q", pgxCfg.Database, pgxCfg.Host, pgxCfg.Port, pgxCfg.User)

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

func readMetasploitConfiguration(filename string) (*pgx.ConnConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filename, err)
	}

	msfConfig := MetasploitDatabaseConfig{}
	err = yaml.Unmarshal(data, &msfConfig)
	if err != nil {
		return nil, fmt.Errorf("parsing YAML in %s: %w", filename, err)
	}

	dbConfig, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}

	dbConfig.Host = msfConfig.Production.Host
	dbConfig.Port = msfConfig.Production.Port
	dbConfig.Database = msfConfig.Production.Database
	dbConfig.User = msfConfig.Production.Username
	dbConfig.Password = msfConfig.Production.Password

	return dbConfig, nil
}
