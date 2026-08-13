package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"saas-superadmin-backend/models"
)

type TenantModule struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Version string `json:"version"`
	Layer   string `json:"layer"`
}

type TenantMonitoring struct {
	ContractVersion string         `json:"contract_version"`
	CheckedAt       string         `json:"checked_at"`
	Status          string         `json:"status"`
	TenantID        string         `json:"tenant_id"`
	Database        string         `json:"database"`
	Company         string         `json:"company"`
	OdooVersion     string         `json:"odoo_version"`
	Users           int            `json:"users"`
	Products        int            `json:"products"`
	Warehouses      int            `json:"warehouses"`
	Modules         []TenantModule `json:"modules"`
}

func FetchTenantMonitoring(tenant models.Tenant) (TenantMonitoring, error) {
	endpoint := fmt.Sprintf(
		"http://odoo_tenant_%d:8069/tenant-agent/v1/health?db=%s",
		tenant.ID,
		url.QueryEscape(tenant.DbName),
	)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return TenantMonitoring{}, err
	}
	if tenant.AgentToken == "" {
		return TenantMonitoring{}, fmt.Errorf("tenant provisioning is incomplete: management agent token is missing")
	}
	req.Header.Set("X-Tenant-Agent-Token", tenant.AgentToken)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TenantMonitoring{}, fmt.Errorf("tenant agent unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return TenantMonitoring{}, fmt.Errorf("tenant agent returned status %d", resp.StatusCode)
	}

	var monitoring TenantMonitoring
	if err := json.NewDecoder(resp.Body).Decode(&monitoring); err != nil {
		return TenantMonitoring{}, fmt.Errorf("invalid tenant agent response: %w", err)
	}
	if monitoring.ContractVersion == "" {
		return TenantMonitoring{}, fmt.Errorf("tenant agent response has no contract version")
	}
	return monitoring, nil
}
