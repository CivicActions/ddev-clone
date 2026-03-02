package clone

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

// configLocalOverrides defines the keys we set in config.local.yaml for clones.
// Using yaml.Node preserves any existing comments and structure in the file.
type configLocalOverrides struct {
	Name string
	// Ports are auto-assigned free ports to avoid conflicts with the source project.
	HostDBPort        string
	HostHTTPSPort     string
	HostWebserverPort string
	HostMailpitPort   string
}

// writeConfigLocalYAML writes or updates .ddev/config.local.yaml with clone-specific
// overrides using yaml.v3's Node API to preserve existing comments and structure.
// If the file already exists (e.g., copied from source), it updates only the fields
// that need to change while keeping everything else intact.
func writeConfigLocalYAML(path string, overrides configLocalOverrides) error {
	var doc yaml.Node

	// Try to read existing file
	data, err := os.ReadFile(path)
	if err == nil && len(data) > 0 {
		if err := yaml.Unmarshal(data, &doc); err != nil {
			// If existing file is malformed, start fresh
			doc = yaml.Node{}
		}
	}

	// Ensure we have a document node with a mapping
	if doc.Kind == 0 {
		// No existing file or empty — create document with mapping
		doc = yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{Kind: yaml.MappingNode},
			},
		}
	}

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return fmt.Errorf("unexpected YAML structure in %s", path)
	}

	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping at top level of %s, got kind %d", path, mapping.Kind)
	}

	// Set the project name override
	setMappingValue(mapping, "name", overrides.Name)

	// Set port overrides — if a value is provided, set it; if empty, remove the key
	// so DDEV uses its defaults without port conflicts being tracked.
	portFields := map[string]string{
		"host_db_port":        overrides.HostDBPort,
		"host_https_port":     overrides.HostHTTPSPort,
		"host_webserver_port": overrides.HostWebserverPort,
		"host_mailpit_port":   overrides.HostMailpitPort,
	}
	for key, value := range portFields {
		if value != "" {
			setMappingValue(mapping, key, value)
		} else {
			removeMappingKey(mapping, key)
		}
	}

	// Marshal back to YAML
	out, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("failed to marshal config.local.yaml: %w", err)
	}

	return os.WriteFile(path, out, 0644)
}

// setMappingValue sets or updates a key-value pair in a YAML mapping node.
// If the key already exists, its value is updated in-place (preserving any
// associated comments). If the key does not exist, it is appended.
func setMappingValue(mapping *yaml.Node, key, value string) {
	// Mapping content is [key1, val1, key2, val2, ...]
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1].Value = value
			mapping.Content[i+1].Tag = "!!str"
			mapping.Content[i+1].Kind = yaml.ScalarNode
			return
		}
	}

	// Key not found — append new key-value pair
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value, Tag: "!!str"},
	)
}

// removeMappingKey removes a key-value pair from a YAML mapping node.
// If the key doesn't exist, this is a no-op.
func removeMappingKey(mapping *yaml.Node, key string) {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			// Remove key and value (2 elements)
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return
		}
	}
}

// getFreePort finds an available TCP port on localhost.
func getFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	return fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port), nil
}

// getFreePorts allocates n unique free TCP ports on localhost.
func getFreePorts(n int) ([]string, error) {
	ports := make([]string, 0, n)
	listeners := make([]*net.TCPListener, 0, n)

	// Hold all listeners open until we've collected all ports
	// to prevent reuse of the same port.
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	for i := 0; i < n; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		listeners = append(listeners, l)
		ports = append(ports, fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port))
	}

	return ports, nil
}
