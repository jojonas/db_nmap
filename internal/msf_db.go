package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

func GetWorkspaceId(ctx context.Context, conn *pgx.Conn, workspaceName string) (int, error) {
	row := conn.QueryRow(ctx, "SELECT id FROM workspaces WHERE name=$1 LIMIT 1", workspaceName)

	var workspaceId int
	err := row.Scan(&workspaceId)
	if err != nil {
		return 0, err
	}

	return workspaceId, nil
}

func InsertHost(ctx context.Context, conn *pgx.Conn, workspaceId int, host NmapHost) (int, error) {
	if !host.HasOpenPorts() {
		log.Debugf("Host %s does not have any open ports, skipping.", host)
		return 0, nil
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	osName := "Unknown"
	if len(host.Os.Osmatch) > 0 {
		osName = host.Os.Osmatch[0].Name
	}

	purpose := "device"
	if len(host.Os.Osclass) > 0 {
		purpose = host.Os.Osclass[0].Type
	}

	allMacs := host.AllMacAddresses()
	preferredMac := ""
	if len(allMacs) > 0 {
		preferredMac = allMacs[0].String()
	}

	allIPs := host.AllIPAddresses()
	preferredIP := ""
	if len(allIPs) > 0 {
		preferredIP = allIPs[0].String()
	}

	allHostnames := host.AllHostnames()
	preferredHostname := ""
	if len(allHostnames) > 0 {
		preferredHostname = allHostnames[0]
	}

	var hostId int
	err = tx.QueryRow(ctx,
		`INSERT INTO hosts (
			created_at, 
			mac,
			address, 
			name, 
			state, 
			os_name, 
			workspace_id, 
			updated_at, 
			purpose 
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (address, workspace_id) DO UPDATE 
		SET updated_at = $1, mac = $2, name = $4, state = $5, os_name = $6, purpose = $9
		RETURNING id`,

		now,               // created_at
		preferredMac,      // mac
		preferredIP,       // address
		preferredHostname, // name
		"alive",           // state
		osName,            // os_name
		workspaceId,       // workspace_id
		now,               // updated_at
		purpose,           // purpose
	).Scan(&hostId)

	if err != nil {
		return 0, fmt.Errorf("inserting host %s: %w", host, err)
	}

	log.Debugf("Inserted/updated host %s.", host)

	openPortCount := 0
	for _, port := range host.Ports.Port {
		err := InsertService(ctx, tx, hostId, port)
		if err != nil {
			return 0, fmt.Errorf("inserting port %s/%d for host %s: %w", port.Protocol, port.Portid, host, err)
		}

		openPortCount++
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commiting transaction: %w", err)
	}

	return openPortCount, nil
}

func InsertService(ctx context.Context, tx pgx.Tx, hostId int, service NmapService) error {
	if service.State.State != "open" {
		return nil
	}

	name := service.Service.Name
	if service.Service.Tunnel != "" {
		name = fmt.Sprintf("%s/%s", service.Service.Tunnel, service.Service.Name)
	}

	now := time.Now()
	_, err := tx.Exec(ctx,
		`INSERT INTO services (
			host_id, 
			created_at, 
			port, 
			proto, 
			state, 
			name, 
			updated_at, 
			info
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (host_id, port, proto) DO UPDATE 
		SET updated_at = $2, state = $5, name = $6, info = $8`,
		hostId,
		now,
		service.Portid,
		service.Protocol,
		service.State.State,
		name,
		now,
		service.Service.Product,
	)

	log.Debugf("Inserted/updated service %s.", service)

	if err != nil {
		return fmt.Errorf("inserting service %s: %w", service, err)
	}

	return nil
}
