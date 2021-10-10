package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

func insertHost(ctx context.Context, conn *pgx.Conn, workspace int, host NmapHost) (int, error) {
	hasOpen := false
	for _, port := range host.Ports.Port {
		if port.State.State == "open" {
			hasOpen = true
			break
		}
	}

	if !hasOpen {
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

	log.Debugf("Inserting/updating host %s (%s)...", host.Address.Addr, host.Hostnames.Hostname.Name)

	var hostId int
	err = tx.QueryRow(ctx,
		`INSERT INTO hosts (
			created_at, 
			address, 
			name, 
			state, 
			os_name, 
			workspace_id, 
			updated_at, 
			purpose 
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (address, workspace_id) DO UPDATE 
		SET updated_at = $1, name = $3, state = $4, os_name = $5, purpose = $8
		RETURNING id`,

		now,                          // created_at
		host.Address.Addr,            // address
		host.Hostnames.Hostname.Name, // name
		"alive",                      // state
		osName,                       // os_name
		workspace,                    // workspace
		now,                          // updated_at
		purpose,                      // purpose
	).Scan(&hostId)

	if err != nil {
		return 0, fmt.Errorf("inserting host %v: %w", host.Address.Addr, err)
	}

	openPortCount := 0
	for _, port := range host.Ports.Port {
		if port.State.State != "open" {
			continue
		}

		openPortCount += 1
		log.Debugf("Inserting/updating service %s:%d (%s) - %s", host.Address.Addr, port.Portid, port.Protocol, port.Service.Name)

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
			port.Portid,
			port.Protocol,
			port.State.State,
			port.Service.Name,
			now,
			port.Service.Product,
		)

		if err != nil {
			return 0, fmt.Errorf("inserting port %s/%d for host %d: %w", port.Protocol, port.Portid, hostId, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commiting transaction: %w", err)
	}

	return openPortCount, nil
}
