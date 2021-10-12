package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	_ "embed"

	"github.com/jojonas/db_nmap/internal"
)

var log = internal.Logger
var binaryPath = "nmap"

func main() {
	if hasArgument(os.Args, "--help") || hasArgument(os.Args, "-h") {
		usage()
	}

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
		conn, err = internal.CheckForReconnect(ctx, conn, 10)
		if err != nil {
			return fmt.Errorf("reconnecting: %w", err)
		}

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

//go:embed usage.txt
var usageTemplate string

func usage() {
	vars := struct {
		Name            string
		ConnString      string
		WorkspaceEnvVar string
		TestedVersions  []string
	}{
		Name:            filepath.Base(os.Args[0]),
		ConnString:      internal.ConnString,
		WorkspaceEnvVar: internal.WorkspaceEnvVar,
		TestedVersions:  internal.TestedVersions,
	}

	tmpl := template.New("usage.txt")
	tmpl = tmpl.Funcs(template.FuncMap{"join": strings.Join})
	tmpl, err := tmpl.Parse(usageTemplate)

	if err != nil {
		log.Errorf("Parsing template usage.txt: %v", err)
		return
	}

	err = tmpl.Execute(os.Stdout, vars)
	if err != nil {
		log.Errorf("Executing template usage.txt: %v", err)
	}
}
