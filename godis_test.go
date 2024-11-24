package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()
	if cli == nil {
		t.Fatal("NewCLI returned nil")
	}

	// Check if configPath is set correctly
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	expectedPath := filepath.Join(homeDir, ".config", "godis")
	if cli.configPath != expectedPath {
		t.Errorf("Expected config path %s, got %s", expectedPath, cli.configPath)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "godis-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test CLI instance
	cli := &CLI{
		configPath: tmpDir,
		configFile: "test-config.yaml",
	}

	// Test loading non-existent config (should create default)
	err = cli.loadConfig()
	if err != nil {
		t.Errorf("Failed to load/create default config: %v", err)
	}

	// Verify default config was created
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Default config file was not created")
	}

	// Test default values
	if cli.config.Username != "godis-bot" {
		t.Errorf("Expected default username 'godis-bot', got %s", cli.config.Username)
	}
}

func TestMergeFlags(t *testing.T) {
	cli := &CLI{
		config: Config{
			Token:     "config-token",
			ChannelID: "config-channel",
			ServerID:  "config-server",
			Username:  "config-user",
		},
		token:     "flag-token",
		channelID: "flag-channel",
		serverID:  "",  // Leave empty to test that config value remains
		username:  "",  // Leave empty to test that config value remains
	}

	cli.mergeFlags()

	// Check that flags override config when present
	if cli.config.Token != "flag-token" {
		t.Errorf("Expected token 'flag-token', got %s", cli.config.Token)
	}
	if cli.config.ChannelID != "flag-channel" {
		t.Errorf("Expected channel 'flag-channel', got %s", cli.config.ChannelID)
	}

	// Check that config values remain when flags are empty
	if cli.config.ServerID != "config-server" {
		t.Errorf("Expected server 'config-server', got %s", cli.config.ServerID)
	}
	if cli.config.Username != "config-user" {
		t.Errorf("Expected username 'config-user', got %s", cli.config.Username)
	}
}

func TestParseFlags(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected CLI
	}{
		{
			name: "Short flags",
			args: []string{"-t", "test-token", "-ch", "test-channel"},
			expected: CLI{
				token:     "test-token",
				channelID: "test-channel",
				configFile: "default.yaml",
			},
		},
		{
			name: "Long flags",
			args: []string{"--token", "test-token", "--channel", "test-channel"},
			expected: CLI{
				token:     "test-token",
				channelID: "test-channel",
				configFile: "default.yaml",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := NewCLI()
			err := cli.parseFlags(tc.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			if cli.token != tc.expected.token {
				t.Errorf("Expected token %s, got %s", tc.expected.token, cli.token)
			}
			if cli.channelID != tc.expected.channelID {
				t.Errorf("Expected channel %s, got %s", tc.expected.channelID, cli.channelID)
			}
			if cli.configFile != tc.expected.configFile {
				t.Errorf("Expected config file %s, got %s", tc.expected.configFile, cli.configFile)
			}
		})
	}
}

func TestTagHandling(t *testing.T) {
	testCases := []struct {
		name        string
		initialTags []string
		flagTags    string
		mode        string
		expected    []string
	}{
		{
			name:        "Merge mode adds new tags",
			initialTags: []string{"one", "two"},
			flagTags:    "two,three",
			mode:        "merge",
			expected:    []string{"one", "two", "three"},
		},
		{
			name:        "Replace mode overwrites tags",
			initialTags: []string{"one", "two"},
			flagTags:    "three,four",
			mode:        "replace",
			expected:    []string{"three", "four"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := NewCLI()
			cli.config.Tags = tc.initialTags
			err := cli.parseFlags([]string{
				"--tags", tc.flagTags,
				"--tag-mode", tc.mode,
			})
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}
			cli.mergeFlags()

			// Compare results (ignoring order)
			if len(cli.config.Tags) != len(tc.expected) {
				t.Errorf("Expected %d tags, got %d", len(tc.expected), len(cli.config.Tags))
			}
			tagMap := make(map[string]bool)
			for _, tag := range cli.config.Tags {
				tagMap[tag] = true
			}
			for _, expectedTag := range tc.expected {
				if !tagMap[expectedTag] {
					t.Errorf("Expected tag %s not found in result", expectedTag)
				}
			}
		})
	}
}

func TestPropertyHandling(t *testing.T) {
	testCases := []struct {
		name            string
		initialProps    map[string]string
		flagProps       string
		mode           string
		expected        map[string]string
	}{
		{
			name:         "Merge mode combines properties",
			initialProps: map[string]string{"one": "1", "two": "2"},
			flagProps:    "two:22;three:3",
			mode:        "merge",
			expected:    map[string]string{"one": "1", "two": "22", "three": "3"},
		},
		{
			name:         "Replace mode overwrites properties",
			initialProps: map[string]string{"one": "1", "two": "2"},
			flagProps:    "three:3;four:4",
			mode:        "replace",
			expected:    map[string]string{"three": "3", "four": "4"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := NewCLI()
			cli.config.Properties = tc.initialProps
			err := cli.parseFlags([]string{
				"--properties", tc.flagProps,
				"--property-mode", tc.mode,
			})
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}
			cli.mergeFlags()

			// Compare results
			if len(cli.config.Properties) != len(tc.expected) {
				t.Errorf("Expected %d properties, got %d", 
					len(tc.expected), len(cli.config.Properties))
			}
			for k, v := range tc.expected {
				if cli.config.Properties[k] != v {
					t.Errorf("Property %s: expected %s, got %s", 
						k, v, cli.config.Properties[k])
				}
			}
		})
	}
}