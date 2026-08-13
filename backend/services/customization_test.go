package services

import (
	"archive/zip"
	"bytes"
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
	_, err := DeployCustomizationArchive("../other-tenant", []byte("not-a-zip"))
	if err == nil || !strings.Contains(err.Error(), "unsafe subdomain") {
		t.Fatalf("expected unsafe subdomain error, got %v", err)
	}
}
