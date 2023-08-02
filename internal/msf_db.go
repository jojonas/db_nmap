package internal

import (
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MsfWorkspace struct {
	Id          int
	Name        string
	Description string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (MsfWorkspace) TableName() string {
	return "workspaces"
}

type MsfHost struct {
	Id          int
	WorkspaceId int
	Address     string

	MAC     string
	Name    string
	State   string
	OSName  string
	Purpose string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (MsfHost) TableName() string {
	return "hosts"
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

func (MsfService) TableName() string {
	return "services"
}

func GetWorkspaceId(db *gorm.DB, workspaceName string) (int, error) {
	var workspace MsfWorkspace

	err := db.Where("name = ?", workspaceName).First(&workspace).Error
	if err != nil {
		return 0, fmt.Errorf("get id for workspace %q: %w", workspaceName, err)
	}

	return workspace.Id, nil
}

func InsertHost(db *gorm.DB, workspaceId int, nmapHost NmapHost) (int, error) {
	if !nmapHost.HasOpenPorts() {
		log.Debugf("Host %s does not have any open ports, skipping.", nmapHost)
		return 0, nil
	}

	now := time.Now()
	openPortCount := 0

	allIPs := nmapHost.AllIPAddresses()
	var preferredIP net.IP
	if len(allIPs) > 0 {
		preferredIP = net.ParseIP(allIPs[0].String())
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var msfHost MsfHost

		msfHost.WorkspaceId = workspaceId
		msfHost.Address = preferredIP.String()

		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("workspace_id = ? AND address = ?", msfHost.WorkspaceId, msfHost.Address).
			FirstOrCreate(&msfHost).
			Error
		if err != nil {
			return fmt.Errorf("query host %v: %w", preferredIP, err)
		}

		msfHost.WorkspaceId = workspaceId
		msfHost.Address = preferredIP.String()

		allMacs := nmapHost.AllMacAddresses()
		if len(allMacs) > 0 {
			msfHost.MAC = allMacs[0].String()
		}

		msfHost.Name = joinHostnames(msfHost.Name, nmapHost.AllHostnames()...)

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

		err = tx.Save(&msfHost).Error
		if err != nil {
			return fmt.Errorf("save host %v: %w", msfHost, err)
		}

		log.Debugf("Inserted/updated host %s.", nmapHost)

		for _, port := range nmapHost.Ports.Port {
			err := InsertService(tx, msfHost.Id, port)
			if err != nil {
				return fmt.Errorf("insert port %s/%d for host %s: %w", port.Protocol, port.Portid, nmapHost, err)
			}

			openPortCount++
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("transaction: %w", err)
	}

	return openPortCount, nil
}

func InsertService(db *gorm.DB, hostId int, service NmapService) error {
	if service.State.State != "open" {
		return nil
	}

	var msfService MsfService

	err := db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"host_id = ? AND proto = ? AND port = ?",
			hostId,
			service.Protocol,
			service.Portid,
		).
		FirstOrCreate(&msfService).
		Error
	if err != nil {
		return fmt.Errorf("query service %v: %w", service, err)
	}

	now := time.Now()

	msfService.HostId = hostId
	msfService.Proto = service.Protocol
	msfService.Port = service.Portid
	msfService.State = service.State.State

	name := service.Service.Name
	if service.Service.Tunnel != "" {
		name = fmt.Sprintf("%s/%s", service.Service.Tunnel, service.Service.Name)
	}
	if name != "" {
		msfService.Name = name
	}

	if service.Service.Product != "" {
		msfService.Info = service.Service.Product

		if service.Service.Version != "" {
			msfService.Info = fmt.Sprintf("%s %s", msfService.Info, service.Service.Version)
		}
	}

	if msfService.CreatedAt.IsZero() {
		msfService.CreatedAt = now
	}

	msfService.UpdatedAt = now

	err = db.Save(&msfService).Error
	if err != nil {
		return fmt.Errorf("save service %v: %w", msfService, err)
	}

	log.Debugf("Inserted/updated service %s.", service)

	return nil
}

// AddNames merges the provided hostnames with the hostnames that are already
// present.
func joinHostnames(previousHostnames string, newHostnames ...string) string {
	var allHostnames []string

	for _, oldHostname := range strings.Split(previousHostnames, ",") {
		allHostnames = append(allHostnames, strings.ToLower(strings.TrimSpace(oldHostname)))
	}

	for _, newHostname := range newHostnames {
		allHostnames = append(allHostnames, strings.ToLower(strings.TrimSpace(newHostname)))
	}

	slices.Sort(allHostnames)

	return strings.Join(slices.Compact(allHostnames), ", ")
}
