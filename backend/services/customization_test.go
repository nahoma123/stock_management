package services

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func customizationZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestValidateCustomizationArchive(t *testing.T) {
	data := customizationZIP(t, map[string]string{
		"customer_stock/__init__.py":     "from . import models\n",
		"customer_stock/__manifest__.py": "{'name': 'Customer Stock'}\n",
	})
	report, err := ValidateCustomizationArchive(data)
	if err != nil {
		t.Fatal(err)
	}
	if report.Module != "customer_stock" || report.Files != 2 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestValidateCustomizationArchiveRejectsUnsafeContent(t *testing.T) {
	tests := map[string]map[string]string{
		"traversal": {"../outside/__manifest__.py": "{}"},
		"multiple modules": {
			"one/__manifest__.py": "{}",
			"two/__manifest__.py": "{}",
		},
		"host process": {
			"unsafe_module/__manifest__.py": "{}",
			"unsafe_module/models.py":       "import subprocess\nsubprocess.run(['id'])",
		},
	}
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ValidateCustomizationArchive(customizationZIP(t, files)); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDeployCustomizationRejectsUnsafeTenantPath(t *testing.T) {
	_, err := StageCustomizationRelease("../other-tenant", 1, []byte("not-a-zip"))
	if err == nil || !strings.Contains(err.Error(), "unsafe subdomain") {
		t.Fatalf("expected unsafe subdomain error, got %v", err)
	}
}

func TestCustomizationReleaseActivationAndRestore(t *testing.T) {
	t.Setenv("TENANT_DATA_ROOT", t.TempDir())
	first := customizationZIP(t, map[string]string{
		"customer_stock/__manifest__.py": "{'name': 'First'}\n",
	})
	second := customizationZIP(t, map[string]string{
		"customer_stock/__manifest__.py": "{'name': 'Second'}\n",
	})

	if _, err := StageCustomizationRelease("acme", 1, first); err != nil {
		t.Fatal(err)
	}
	if _, err := StageCustomizationRelease("acme", 2, second); err != nil {
		t.Fatal(err)
	}
	if _, err := ActivateCustomizationRelease("acme", "customer_stock", 1); err != nil {
		t.Fatal(err)
	}
	root := tenantCustomizationRoot("acme")
	activePath := filepath.Join(root, "customer_stock")
	target, err := os.Readlink(activePath)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(target) {
		t.Fatalf("active release link must be portable, got absolute target %q", target)
	}

	activation, err := ActivateCustomizationRelease("acme", "customer_stock", 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := RestoreCustomizationActivation("acme", activation); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join(activePath, "__manifest__.py"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), "First") {
		t.Fatalf("restore did not reactivate the first release: %s", manifest)
	}
}

func TestCustomizationReleaseRejectsUnmanagedActiveModule(t *testing.T) {
	t.Setenv("TENANT_DATA_ROOT", t.TempDir())
	root := tenantCustomizationRoot("new-tenant")
	activeModule := filepath.Join(root, "customer_stock")
	if err := os.MkdirAll(activeModule, 0755); err != nil {
		t.Fatal(err)
	}
	data := customizationZIP(t, map[string]string{
		"customer_stock/__manifest__.py": "{'name': 'Release'}\n",
	})
	if _, err := StageCustomizationRelease("new-tenant", 1, data); err != nil {
		t.Fatal(err)
	}
	if _, err := ActivateCustomizationRelease("new-tenant", "customer_stock", 1); err == nil || !strings.Contains(err.Error(), "not managed") {
		t.Fatalf("expected unmanaged module error, got %v", err)
	}
}
