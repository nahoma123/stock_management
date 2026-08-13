package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const MaxCustomizationSize = 25 << 20

var (
	moduleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
	tenantPathPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)
	blockedCode       = []string{"/var/run/docker.sock", "os.system(", "subprocess.", "pty.spawn("}
)

type CustomizationReport struct {
	Module   string   `json:"module"`
	Files    int      `json:"files"`
	Size     int64    `json:"size"`
	Warnings []string `json:"warnings"`
}

func ValidateCustomizationArchive(data []byte) (CustomizationReport, error) {
	if len(data) == 0 {
		return CustomizationReport{}, fmt.Errorf("archive is empty")
	}
	if len(data) > MaxCustomizationSize {
		return CustomizationReport{}, fmt.Errorf("archive exceeds the 25 MB limit")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return CustomizationReport{}, fmt.Errorf("invalid ZIP archive: %w", err)
	}

	topLevels := map[string]bool{}
	manifestFound := false
	var totalSize int64
	for _, file := range reader.File {
		if strings.Contains(file.Name, "\\") {
			return CustomizationReport{}, fmt.Errorf("archive paths must use forward slashes: %s", file.Name)
		}
		cleanName := filepath.ToSlash(filepath.Clean(file.Name))
		if cleanName == "." || strings.HasPrefix(cleanName, "../") || strings.HasPrefix(cleanName, "/") {
			return CustomizationReport{}, fmt.Errorf("archive contains an unsafe path: %s", file.Name)
		}
		parts := strings.Split(cleanName, "/")
		topLevels[parts[0]] = true
		if file.Mode()&os.ModeSymlink != 0 {
			return CustomizationReport{}, fmt.Errorf("symbolic links are not allowed: %s", file.Name)
		}
		if file.UncompressedSize64 > uint64(MaxCustomizationSize) {
			return CustomizationReport{}, fmt.Errorf("file exceeds the 25 MB limit: %s", file.Name)
		}
		totalSize += int64(file.UncompressedSize64)
		if totalSize > MaxCustomizationSize {
			return CustomizationReport{}, fmt.Errorf("uncompressed archive exceeds the 25 MB limit")
		}
	}
	if len(topLevels) != 1 {
		return CustomizationReport{}, fmt.Errorf("archive must contain exactly one top-level module directory")
	}
	var moduleName string
	for name := range topLevels {
		moduleName = name
	}
	if !moduleNamePattern.MatchString(moduleName) {
		return CustomizationReport{}, fmt.Errorf("invalid Odoo module directory name %q", moduleName)
	}

	platformModule := filepath.Join("/mnt/extra-addons", moduleName, "__manifest__.py")
	if _, err := os.Stat(platformModule); err == nil {
		return CustomizationReport{}, fmt.Errorf("module name %q belongs to the platform layer", moduleName)
	}

	for _, file := range reader.File {
		cleanName := filepath.ToSlash(filepath.Clean(file.Name))
		if cleanName == moduleName+"/__manifest__.py" {
			manifestFound = true
		}
		if !strings.HasSuffix(cleanName, ".py") || file.FileInfo().IsDir() {
			continue
		}
		content, err := readZipFile(file, 2<<20)
		if err != nil {
			return CustomizationReport{}, err
		}
		for _, blocked := range blockedCode {
			if strings.Contains(string(content), blocked) {
				return CustomizationReport{}, fmt.Errorf("%s contains blocked host-process access: %s", cleanName, blocked)
			}
		}
	}
	if !manifestFound {
		return CustomizationReport{}, fmt.Errorf("module is missing __manifest__.py")
	}

	warnings := []string{
		"Python modules execute with the tenant Odoo process and require operator review.",
		"Deployment updates the tenant addon layer but does not automatically install the module.",
	}
	sort.Strings(warnings)
	return CustomizationReport{Module: moduleName, Files: len(reader.File), Size: totalSize, Warnings: warnings}, nil
}

func DeployCustomizationArchive(tenantSubdomain string, data []byte) (CustomizationReport, error) {
	if !tenantPathPattern.MatchString(tenantSubdomain) {
		return CustomizationReport{}, fmt.Errorf("tenant has an unsafe subdomain")
	}
	report, err := ValidateCustomizationArchive(data)
	if err != nil {
		return CustomizationReport{}, err
	}
	root := filepath.Join("/app/tenants", tenantSubdomain, "custom_addons")
	if err := os.MkdirAll(root, 0755); err != nil {
		return CustomizationReport{}, err
	}
	tempRoot, err := os.MkdirTemp(root, ".deployment-")
	if err != nil {
		return CustomizationReport{}, err
	}
	defer os.RemoveAll(tempRoot)

	reader, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	for _, file := range reader.File {
		target := filepath.Join(tempRoot, filepath.FromSlash(filepath.Clean(file.Name)))
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return CustomizationReport{}, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return CustomizationReport{}, err
		}
		source, err := file.Open()
		if err != nil {
			return CustomizationReport{}, err
		}
		destination, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			source.Close()
			return CustomizationReport{}, err
		}
		_, copyErr := io.Copy(destination, source)
		source.Close()
		destination.Close()
		if copyErr != nil {
			return CustomizationReport{}, copyErr
		}
	}

	sourceModule := filepath.Join(tempRoot, report.Module)
	targetModule := filepath.Join(root, report.Module)
	backupModule := filepath.Join(root, ".backups", report.Module)
	if err := os.MkdirAll(filepath.Dir(backupModule), 0755); err != nil {
		return CustomizationReport{}, err
	}
	os.RemoveAll(backupModule)
	if _, err := os.Stat(targetModule); err == nil {
		if err := os.Rename(targetModule, backupModule); err != nil {
			return CustomizationReport{}, err
		}
	}
	if err := os.Rename(sourceModule, targetModule); err != nil {
		if _, backupErr := os.Stat(backupModule); backupErr == nil {
			_ = os.Rename(backupModule, targetModule)
		}
		return CustomizationReport{}, err
	}
	return report, nil
}

func readZipFile(file *zip.File, limit int64) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	content, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > limit {
		return nil, fmt.Errorf("file is too large to inspect: %s", file.Name)
	}
	return content, nil
}
