package internal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v4"
)

type MsfHost struct {
	Id          int
	WorkspaceId int
	Address     net.IP

	MAC     string
	Name    string
	State   string
	OSName  string
	Purpose string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type MsfService struct {
	Id        int
	HostId    int
	CreatedAt time.Time
	Port      int
	Proto     string
	State     string
	Name      string
	UpdatedAt time.Time
	Info      string
}

func GetWorkspaceId(ctx context.Context, conn *pgx.Conn, workspaceName string) (int, error) {
	row := conn.QueryRow(ctx, "SELECT id FROM workspaces WHERE name=$1 LIMIT 1", workspaceName)

	var workspaceId int
	err := row.Scan(&workspaceId)
	if err != nil {
		return 0, err
	}

	return workspaceId, nil
}

func InsertHost(ctx context.Context, conn *pgx.Conn, workspaceId int, nmapHost NmapHost) (int, error) {
	if !nmapHost.HasOpenPorts() {
		log.Debugf("Host %s does not have any open ports, skipping.", nmapHost)
		return 0, nil
	}

	now := time.Now()

	allIPs := nmapHost.AllIPAddresses()
	var preferredIP net.IP
	if len(allIPs) > 0 {
		preferredIP = net.ParseIP(allIPs[0].String())
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	msfHost := MsfHost{}

	err = tx.QueryRow(ctx,
		`SELECT 
				id, 
				workspace_id, 
				address, 
				mac, 
				name, 
				state, 
				os_name, 
				purpose,
				created_at, 
				updated_at
			FROM hosts 
			WHERE workspace_id=$1 AND address=$2
			LIMIT 1
			FOR UPDATE`,
		workspaceId,
		preferredIP.String(),
	).Scan(
		&msfHost.Id,
		&msfHost.WorkspaceId,
		&msfHost.Address,
		&msfHost.MAC,
		&msfHost.Name,
		&msfHost.State,
		&msfHost.OSName,
		&msfHost.Purpose,
		&msfHost.CreatedAt,
		&msfHost.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Debugf("host %v not found in database", preferredIP)
	} else if err != nil {
		return 0, fmt.Errorf("query host %v: %w", preferredIP, err)
	}

	msfHost.WorkspaceId = workspaceId
	msfHost.Address = preferredIP

	allMacs := nmapHost.AllMacAddresses()
	if len(allMacs) > 0 {
		msfHost.MAC = allMacs[0].String()
	}

	allHostnames := nmapHost.AllHostnames()
	if len(allHostnames) > 0 {
		msfHost.Name = allHostnames[0]
	}

	msfHost.State = "alive"

	if len(nmapHost.Os.Osmatch) > 0 {
		msfHost.OSName = nmapHost.Os.Osmatch[0].Name
	}

	if msfHost.Purpose == "" {
		msfHost.Purpose = "device"
	}

	if len(nmapHost.Os.Osclass) > 0 {
		msfHost.Purpose = nmapHost.Os.Osclass[0].Type
	}

	if msfHost.CreatedAt.IsZero() {
		msfHost.CreatedAt = now
	}

	msfHost.UpdatedAt = now

	err = tx.QueryRow(ctx,
		`INSERT INTO hosts (
			workspace_id, 
			address, 
			mac, 
			name, 
			state, 
			os_name, 
			purpose,
			created_at, 
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (workspace_id, address) DO UPDATE SET 
			mac = excluded.mac, 
			name = excluded.name, 
			state = excluded.state, 
			os_name = excluded.os_name, 
			purpose = excluded.purpose, 
			created_at = excluded.created_at, 
			updated_at = excluded.updated_at
		RETURNING
			id,
			workspace_id, 
			address, 
			mac, 
			name, 
			state, 
			os_name, 
			purpose,
			created_at, 
			updated_at
		`,
		msfHost.WorkspaceId,
		msfHost.Address.String(), // need to call String(), otherwise IPv4 addresses are inserted as IPv6
		msfHost.MAC,
		msfHost.Name,
		msfHost.State,
		msfHost.OSName,
		msfHost.Purpose,
		msfHost.CreatedAt,
		msfHost.UpdatedAt,
	).Scan(
		&msfHost.Id,
		&msfHost.WorkspaceId,
		&msfHost.Address,
		&msfHost.MAC,
		&msfHost.Name,
		&msfHost.State,
		&msfHost.OSName,
		&msfHost.Purpose,
		&msfHost.CreatedAt,
		&msfHost.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("upsert host %s: %w", nmapHost, err)
	}

	log.Debugf("Inserted/updated host %s.", nmapHost)

	openPortCount := 0
	for _, port := range nmapHost.Ports.Port {
		err := InsertService(ctx, tx, msfHost.Id, port)
		if err != nil {
			return 0, fmt.Errorf("insert port %s/%d for host %s: %w", port.Protocol, port.Portid, nmapHost, err)
		}

		openPortCount++
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
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
		return fmt.Errorf("insert service %s: %w", service, err)
	}

	return nil
}
