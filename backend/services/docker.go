package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"saas-superadmin-backend/models"
)

type RestartPolicy struct {
	Name string `json:"Name"`
}

type HostConfig struct {
	NetworkMode   string        `json:"NetworkMode"`
	Binds         []string      `json:"Binds"`
	RestartPolicy RestartPolicy `json:"RestartPolicy,omitempty"`
}

func addonBinds(hostVolumePath string) []string {
	hostProjectPath := GetEnv("HOST_PROJECT_PATH", "/home/nahom/Desktop/project/odoo_project")
	return []string{
		fmt.Sprintf("%s:/mnt/tenant-addons", hostVolumePath),
		fmt.Sprintf("%s/custom_addons:/mnt/platform-addons:ro", hostProjectPath),
	}
}

type ContainerConfig struct {
	Image      string            `json:"Image"`
	Cmd        []string          `json:"Cmd"`
	Env        []string          `json:"Env"`
	HostConfig HostConfig        `json:"HostConfig"`
	Labels     map[string]string `json:"Labels"`
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func CallDockerAPI(method, path string, payload interface{}) ([]byte, int, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	var reqBody io.Reader
	if payload != nil {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, "http://localhost"+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func EnsureDummyModuleExists(subdomain string) error {
	tenantPath := filepath.Join("/app/tenants", subdomain, "custom_addons")
	if err := os.MkdirAll(tenantPath, 0755); err != nil {
		return err
	}

	dummyModulePath := filepath.Join(tenantPath, "dummy_module")
	if err := os.MkdirAll(dummyModulePath, 0755); err != nil {
		return err
	}

	initPyPath := filepath.Join(dummyModulePath, "__init__.py")
	if _, err := os.Stat(initPyPath); os.IsNotExist(err) {
		if err := os.WriteFile(initPyPath, []byte(""), 0644); err != nil {
			return err
		}
	}

	manifestPyPath := filepath.Join(dummyModulePath, "__manifest__.py")
	if _, err := os.Stat(manifestPyPath); os.IsNotExist(err) {
		manifestContent := `{
    'name': 'Dummy Module',
    'version': '1.0',
    'category': 'Hidden',
    'summary': 'Dummy module for path validation',
    'depends': ['base'],
    'installable': True,
}`
		if err := os.WriteFile(manifestPyPath, []byte(manifestContent), 0644); err != nil {
			return err
		}
	}
	return nil
}

func DockerCreateAndStartInitContainer(tenant models.Tenant, dbHost, dbPort, dbUser, dbPassword string, hostVolumePath string) error {
	dockerNetwork := GetEnv("DOCKER_NETWORK", "saas_net")
	containerName := fmt.Sprintf("odoo_init_%d", tenant.ID)

	CallDockerAPI("DELETE", fmt.Sprintf("/containers/%s?force=true", containerName), nil)

	config := ContainerConfig{
		Image: "odoo_project-odoo:latest",
		Cmd: []string{
			"odoo", "--config=/dev/null", "--database", tenant.DbName,
			"--db_host", dbHost, "--db_port", dbPort, "--db_user", dbUser, "--db_password", dbPassword,
			"--addons-path=/mnt/tenant-addons,/mnt/platform-addons,/usr/lib/python3/dist-packages/odoo/addons",
			"--init", "base,web,sale_management,stock,daily_sales_report,initial_data_import",
			"--stop-after-init",
		},
		HostConfig: HostConfig{
			NetworkMode: dockerNetwork,
			Binds:       addonBinds(hostVolumePath),
		},
	}

	UpdateLogAndBroadcast(tenant.ID, "Creating transient initialization container...\n")
	resp, status, err := CallDockerAPI("POST", fmt.Sprintf("/containers/create?name=%s", containerName), config)
	if err != nil {
		return fmt.Errorf("failed to create init container: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("failed to create init container (status %d): %s", status, string(resp))
	}

	UpdateLogAndBroadcast(tenant.ID, "Starting initialization container...\n")
	resp, status, err = CallDockerAPI("POST", fmt.Sprintf("/containers/%s/start", containerName), nil)
	if err != nil {
		return fmt.Errorf("failed to start init container: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("failed to start init container (status %d): %s", status, string(resp))
	}

	UpdateLogAndBroadcast(tenant.ID, "Waiting for database initialization to complete (this may take 1-2 minutes)...\n")
	resp, status, err = CallDockerAPI("POST", fmt.Sprintf("/containers/%s/wait", containerName), nil)
	if err != nil {
		return fmt.Errorf("failed to wait for init container: %w", err)
	}

	logResp, _, _ := CallDockerAPI("GET", fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true", containerName), nil)
	cleanLogs := ParseDockerLogs(logResp)
	UpdateLogAndBroadcast(tenant.ID, cleanLogs)

	CallDockerAPI("DELETE", fmt.Sprintf("/containers/%s?force=true", containerName), nil)

	return nil
}

func DockerCreateAndStartDaemonContainer(tenant models.Tenant, dbHost, dbPort, dbUser, dbPassword string, hostVolumePath string) error {
	dockerNetwork := GetEnv("DOCKER_NETWORK", "saas_net")
	containerName := fmt.Sprintf("odoo_tenant_%d", tenant.ID)

	CallDockerAPI("DELETE", fmt.Sprintf("/containers/%s?force=true", containerName), nil)

	config := ContainerConfig{
		Image: "odoo_project-odoo:latest",
		Cmd: []string{
			"odoo", "-r", dbUser, "-w", dbPassword, "--db_host", dbHost, "--db_port", dbPort, "--database", tenant.DbName,
			"--addons-path=/mnt/tenant-addons,/mnt/platform-addons,/usr/lib/python3/dist-packages/odoo/addons",
		},
		Env: []string{
			fmt.Sprintf("DB_HOST=%s", dbHost),
			fmt.Sprintf("DB_PORT=%s", dbPort),
			fmt.Sprintf("DB_USER=%s", dbUser),
			fmt.Sprintf("DB_PASSWORD=%s", dbPassword),
		},
		HostConfig: HostConfig{
			NetworkMode: dockerNetwork,
			Binds:       addonBinds(hostVolumePath),
			RestartPolicy: RestartPolicy{
				Name: "unless-stopped",
			},
		},
		Labels: map[string]string{
			"traefik.enable": "true",
			fmt.Sprintf("traefik.http.routers.%s.rule", containerName):                      fmt.Sprintf("Host(`%s.localhost`)", tenant.Subdomain),
			fmt.Sprintf("traefik.http.routers.%s.entrypoints", containerName):               "web",
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", containerName): "8069",
		},
	}

	UpdateLogAndBroadcast(tenant.ID, "Creating tenant container daemon...\n")
	resp, status, err := CallDockerAPI("POST", fmt.Sprintf("/containers/create?name=%s", containerName), config)
	if err != nil {
		return fmt.Errorf("failed to create daemon container: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("failed to create daemon container (status %d): %s", status, string(resp))
	}

	UpdateLogAndBroadcast(tenant.ID, "Starting tenant container daemon...\n")
	resp, status, err = CallDockerAPI("POST", fmt.Sprintf("/containers/%s/start", containerName), nil)
	if err != nil {
		return fmt.Errorf("failed to start daemon container: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("failed to start daemon container (status %d): %s", status, string(resp))
	}

	return nil
}

func ParseDockerLogs(raw []byte) string {
	var builder strings.Builder
	i := 0
	for i < len(raw) {
		if i+8 > len(raw) {
			break
		}
		size := int(raw[i+4])<<24 | int(raw[i+5])<<16 | int(raw[i+6])<<8 | int(raw[i+7])
		i += 8
		if i+size > len(raw) {
			builder.Write(raw[i:])
			break
		}
		builder.Write(raw[i : i+size])
		i += size
	}
	return builder.String()
}
