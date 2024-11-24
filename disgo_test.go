package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestStdinHandlingWithPassthrough(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
			t.Fatalf("Failed to create pipe: %v", err)
	}

	// Create a pipe for capturing stdout
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
			t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// Save original stdin and stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
			os.Stdin = oldStdin
			os.Stdout = oldStdout
	}()

	// Set the pipes
	os.Stdin = r
	os.Stdout = stdoutW

	// Test data to write
	testData := []byte("test input data\n")

	// Create a channel to collect output
	outputChan := make(chan []byte)
	
	// Start goroutine to read from stdout pipe
	go func() {
			var buf bytes.Buffer
			io.Copy(&buf, stdoutR)
			outputChan <- buf.Bytes()
	}()

	// Write test data and close pipe
	go func() {
			w.Write(testData)
			w.Close()
	}()

	// Run with passthrough enabled
	cli := NewCLI()
	cli.parseFlags([]string{"--passthrough"})
	cli.readStdin()
	cli.config.Passthrough = true

	// Write to stdout if passthrough enabled
	if cli.config.Passthrough && len(cli.stdinData) > 0 {
			os.Stdout.Write(cli.stdinData)
	}

	// Close stdout pipe to signal completion
	stdoutW.Close()

	// Get the captured output
	output := <-outputChan

	// Verify the output matches input when passthrough is enabled
	if !bytes.Equal(output, testData) {
			t.Errorf("Passthrough output mismatch:\nexpected: %q\ngot: %q", testData, output)
	}
}

func TestStdinHandlingWithoutPassthrough(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
			t.Fatalf("Failed to create pipe: %v", err)
	}

	// Create a pipe for capturing stdout
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
			t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// Save original stdin and stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
			os.Stdin = oldStdin
			os.Stdout = oldStdout
	}()

	// Set the pipes
	os.Stdin = r
	os.Stdout = stdoutW

	// Test data to write
	testData := []byte("test input data\n")

	// Create a channel to collect output
	outputChan := make(chan []byte)
	
	// Start goroutine to read from stdout pipe
	go func() {
			var buf bytes.Buffer
			io.Copy(&buf, stdoutR)
			outputChan <- buf.Bytes()
	}()

	// Write test data and close pipe
	go func() {
			w.Write(testData)
			w.Close()
	}()

	// Run without passthrough
	cli := NewCLI()
	cli.readStdin()
	
	// Should not write to stdout as passthrough is disabled
	if cli.config.Passthrough && len(cli.stdinData) > 0 {
			os.Stdout.Write(cli.stdinData)
	}

	// Close stdout pipe to signal completion
	stdoutW.Close()

	// Get the captured output
	output := <-outputChan

	// Verify no output when passthrough is disabled
	if len(output) > 0 {
			t.Errorf("Expected no output without passthrough, got: %q", output)
	}

	// Verify data was still read
	if !bytes.Equal(cli.stdinData, testData) {
			t.Errorf("Stdin data mismatch:\nexpected: %q\ngot: %q", testData, cli.stdinData)
	}
}
func TestEffectiveMaxMessageSize(t *testing.T) {
	testCases := []struct {
			name           string
			configSize     int
			expectedSize   int
	}{
			{
					name:         "Zero size falls back to default",
					configSize:   0,
					expectedSize: DefaultMaxMessageSize,
			},
			{
					name:         "Negative size falls back to default",
					configSize:   -1,
					expectedSize: DefaultMaxMessageSize,
			},
			{
					name:         "Valid size is used",
					configSize:   1000,
					expectedSize: 1000,
			},
			{
					name:         "Default size when not set",
					configSize:   0,
					expectedSize: DefaultMaxMessageSize,
			},
	}

	for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
					cli := NewCLI()
					cli.config.MaxMessageSize = tc.configSize

					size := cli.getEffectiveMaxMessageSize()
					if size != tc.expectedSize {
							t.Errorf("Expected size %d, got %d", tc.expectedSize, size)
					}

					// Test message splitting with this size
					longMessage := strings.Repeat("a", tc.expectedSize + 100)
					messages := cli.splitMessage(longMessage)
					
					// Verify no message exceeds the effective max size
					for i, msg := range messages {
							if len(msg) > tc.expectedSize {
									t.Errorf("Message part %d exceeds max size: %d > %d", 
											i, len(msg), tc.expectedSize)
							}
					}
			})
	}
}

func TestMessageModeHandling(t *testing.T) {
	testCases := []struct {
			name        string
			mode        string
			content     string
			expectParts int
	}{
			{
					name:        "Invalid mode defaults to truncate",
					mode:        "invalid_mode",
					content:     strings.Repeat("a", DefaultMaxMessageSize * 2),
					expectParts: 1,
			},
			{
					name:        "Empty mode defaults to truncate",
					mode:        "",
					content:     strings.Repeat("a", DefaultMaxMessageSize * 2),
					expectParts: 1,
			},
	}

	for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
					cli := NewCLI()
					cli.config.MessageMode = tc.mode
					cli.config.MaxMessageSize = DefaultMaxMessageSize

					messages := cli.splitMessage(tc.content)
					if len(messages) != tc.expectParts {
							t.Errorf("Expected %d parts, got %d", tc.expectParts, len(messages))
					}

					// Verify size limits
					for i, msg := range messages {
							if len(msg) > DefaultMaxMessageSize {
									t.Errorf("Message part %d exceeds max size: %d > %d", 
											i, len(msg), DefaultMaxMessageSize)
							}
					}
			})
	}
}
