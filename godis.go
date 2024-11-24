package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration options
type Config struct {
	Token     string `yaml:"token"`
	ChannelID string `yaml:"channel_id"`
	ServerID  string `yaml:"server_id"`
	Username  string `yaml:"username"`
}

type CLI struct {
	config      Config
	configFile  string
	configPath  string
	token       string
	channelID   string
	serverID    string
	username    string
}

func NewCLI() *CLI {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	return &CLI{
		configPath: filepath.Join(homeDir, ".config", "godis"),
	}
}

func (c *CLI) parseFlags() {
	// Define flags with long and short versions
	flag.StringVar(&c.configFile, "config", "default.yaml", "Config file to use")
	flag.StringVar(&c.configFile, "c", "default.yaml", "Config file to use (shorthand)")
	
	flag.StringVar(&c.token, "token", "", "Discord bot token")
	flag.StringVar(&c.token, "t", "", "Discord bot token (shorthand)")
	
	flag.StringVar(&c.channelID, "channel", "", "Discord channel ID")
	flag.StringVar(&c.channelID, "ch", "", "Discord channel ID (shorthand)")
	
	flag.StringVar(&c.serverID, "server", "", "Discord server ID")
	flag.StringVar(&c.serverID, "s", "", "Discord server ID (shorthand)")
	
	flag.StringVar(&c.username, "username", "", "Bot username")
	flag.StringVar(&c.username, "u", "", "Bot username (shorthand)")

	flag.Parse()
}

func (c *CLI) loadConfig() error {
	// Ensure config directory exists
	err := os.MkdirAll(c.configPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load config file
	configFile := filepath.Join(c.configPath, c.configFile)
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if it doesn't exist
			return c.createDefaultConfig(configFile)
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	err = yaml.Unmarshal(data, &c.config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (c *CLI) createDefaultConfig(configFile string) error {
	defaultConfig := Config{
		Token:     "",
		ChannelID: "",
		ServerID:  "",
		Username:  "godis-bot",
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	err = ioutil.WriteFile(configFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	c.config = defaultConfig
	return nil
}

func (c *CLI) mergeFlags() {
	// Override config with command line flags if provided
	if c.token != "" {
		c.config.Token = c.token
	}
	if c.channelID != "" {
		c.config.ChannelID = c.channelID
	}
	if c.serverID != "" {
		c.config.ServerID = c.serverID
	}
	if c.username != "" {
		c.config.Username = c.username
	}
}

func main() {
	cli := NewCLI()
	cli.parseFlags()

	err := cli.loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	cli.mergeFlags()

	// TODO: Implement Discord message handling
	// For now, just print the configuration
	fmt.Printf("Using configuration:\n")
	fmt.Printf("Token: %s\n", cli.config.Token)
	fmt.Printf("Channel ID: %s\n", cli.config.ChannelID)
	fmt.Printf("Server ID: %s\n", cli.config.ServerID)
	fmt.Printf("Username: %s\n", cli.config.Username)
}