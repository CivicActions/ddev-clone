package clone

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteConfigLocalYAML_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.local.yaml")

	err := writeConfigLocalYAML(path, configLocalOverrides{
		Name:              "my-clone",
		HostDBPort:        "33061",
		HostHTTPSPort:     "44301",
		HostWebserverPort: "8081",
		HostMailpitPort:   "8028",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)
	// Verify all expected keys are present
	for _, key := range []string{"name: my-clone", "host_db_port:", "host_https_port:", "host_webserver_port:", "host_mailpit_port:"} {
		if !strings.Contains(content, key) {
			t.Errorf("expected %q in output, got:\n%s", key, content)
		}
	}
}

func TestWriteConfigLocalYAML_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.local.yaml")

	// Write an existing file with comments and extra fields
	existing := `# This is a local config override
name: old-project
# Keep the DB port at default
host_db_port: "3306"
# Custom setting that should be preserved
php_version: "8.2"
# Another comment
xdebug_enabled: true
`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	err := writeConfigLocalYAML(path, configLocalOverrides{
		Name:              "new-clone",
		HostDBPort:        "33099",
		HostHTTPSPort:     "44399",
		HostWebserverPort: "8099",
		HostMailpitPort:   "8099",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)

	// Verify overrides were applied
	if !strings.Contains(content, "name: new-clone") {
		t.Errorf("expected updated name, got:\n%s", content)
	}

	// Verify comments are preserved
	if !strings.Contains(content, "# This is a local config override") {
		t.Errorf("expected top comment preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "# Keep the DB port at default") {
		t.Errorf("expected DB port comment preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "# Custom setting that should be preserved") {
		t.Errorf("expected custom comment preserved, got:\n%s", content)
	}

	// Verify non-override fields are preserved
	if !strings.Contains(content, "php_version:") {
		t.Errorf("expected php_version preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "xdebug_enabled:") {
		t.Errorf("expected xdebug_enabled preserved, got:\n%s", content)
	}

	// Verify old name was replaced, not duplicated
	if strings.Contains(content, "old-project") {
		t.Errorf("expected old name to be replaced, got:\n%s", content)
	}

	// Verify DB port was updated, not at old value
	if strings.Contains(content, "3306") {
		t.Errorf("expected DB port updated from 3306, got:\n%s", content)
	}

	// Verify new keys were added
	if !strings.Contains(content, "host_https_port:") {
		t.Errorf("expected host_https_port added, got:\n%s", content)
	}
	if !strings.Contains(content, "host_webserver_port:") {
		t.Errorf("expected host_webserver_port added, got:\n%s", content)
	}
	if !strings.Contains(content, "host_mailpit_port:") {
		t.Errorf("expected host_mailpit_port added, got:\n%s", content)
	}
}

func TestWriteConfigLocalYAML_PreservesExistingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.local.yaml")

	// Existing file with many custom fields
	existing := `name: source-project
router_http_port: "8080"
router_https_port: "8443"
additional_hostnames:
  - mysite.local
  - api.mysite.local
`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	err := writeConfigLocalYAML(path, configLocalOverrides{
		Name:              "clone-project",
		HostDBPort:        "33088",
		HostHTTPSPort:     "44388",
		HostWebserverPort: "8088",
		HostMailpitPort:   "8088",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)

	// Name should be updated
	if !strings.Contains(content, "name: clone-project") {
		t.Errorf("expected updated name, got:\n%s", content)
	}

	// Existing fields should be preserved
	if !strings.Contains(content, "router_http_port:") {
		t.Errorf("expected router_http_port preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "router_https_port:") {
		t.Errorf("expected router_https_port preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "additional_hostnames:") {
		t.Errorf("expected additional_hostnames preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "mysite.local") {
		t.Errorf("expected hostname entries preserved, got:\n%s", content)
	}
}

func TestWriteConfigLocalYAML_EmptyExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.local.yaml")

	// Empty file
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	err := writeConfigLocalYAML(path, configLocalOverrides{
		Name:              "empty-test",
		HostDBPort:        "33077",
		HostHTTPSPort:     "44377",
		HostWebserverPort: "8077",
		HostMailpitPort:   "8078",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "name: empty-test") {
		t.Errorf("expected name in output, got:\n%s", content)
	}
}

func TestWriteConfigLocalYAML_InlineComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.local.yaml")

	// File with inline comments
	existing := `name: my-project # project name
host_db_port: "3306" # default DB port
custom_field: value # keep this
`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	err := writeConfigLocalYAML(path, configLocalOverrides{
		Name:              "updated-project",
		HostDBPort:        "33066",
		HostHTTPSPort:     "44366",
		HostWebserverPort: "8066",
		HostMailpitPort:   "8067",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)

	// Updated values
	if !strings.Contains(content, "name: updated-project") {
		t.Errorf("expected updated name, got:\n%s", content)
	}

	// Custom field preserved
	if !strings.Contains(content, "custom_field:") {
		t.Errorf("expected custom_field preserved, got:\n%s", content)
	}

	// Inline comment on custom_field should be preserved
	if !strings.Contains(content, "# keep this") {
		t.Errorf("expected inline comment preserved, got:\n%s", content)
	}
}

func TestWriteConfigLocalYAML_EmptyPortRemovesKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.local.yaml")

	// Existing file with port fields (e.g., copied from source)
	existing := `name: source-project
host_db_port: "3306"
host_https_port: "443"
host_webserver_port: "80"
host_mailpit_port: "8027"
php_version: "8.2"
`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	// Empty port values should remove the keys from YAML
	err := writeConfigLocalYAML(path, configLocalOverrides{
		Name: "clone-project",
		// Empty ports — should be removed
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)

	// Name should be updated
	if !strings.Contains(content, "name: clone-project") {
		t.Errorf("expected updated name, got:\n%s", content)
	}

	// Port keys should be removed
	if strings.Contains(content, "host_db_port") {
		t.Errorf("expected host_db_port to be removed, got:\n%s", content)
	}
	if strings.Contains(content, "host_https_port") {
		t.Errorf("expected host_https_port to be removed, got:\n%s", content)
	}
	if strings.Contains(content, "host_webserver_port") {
		t.Errorf("expected host_webserver_port to be removed, got:\n%s", content)
	}
	if strings.Contains(content, "host_mailpit_port") {
		t.Errorf("expected host_mailpit_port to be removed, got:\n%s", content)
	}

	// Non-port fields preserved
	if !strings.Contains(content, "php_version:") {
		t.Errorf("expected php_version preserved, got:\n%s", content)
	}
}

func TestGetFreePorts(t *testing.T) {
	ports, err := getFreePorts(4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ports) != 4 {
		t.Fatalf("expected 4 ports, got %d", len(ports))
	}

	// All ports should be unique
	seen := make(map[string]bool)
	for _, p := range ports {
		if seen[p] {
			t.Errorf("duplicate port: %s", p)
		}
		seen[p] = true

		// Port should be a reasonable number (not 0, not super low)
		if p == "0" || p == "" {
			t.Errorf("invalid port: %q", p)
		}
	}
}
