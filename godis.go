package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type MergeMode string
const (
	ModeReplace MergeMode = "replace"
	ModeMerge   MergeMode = "merge"
)

// Config holds all configuration options
type Config struct {
	Token     string `yaml:"token"`
	ChannelID string `yaml:"channel_id"`
	ServerID  string `yaml:"server_id"`
	Username  string `yaml:"username"`
	Tags      []string `yaml:"tags"`
	Properties map[string]string `yaml:"properties"`
}

type CLI struct {
	config      Config
	configFile  string
	configPath  string
	token       string
	channelID   string
	serverID    string
	username    string
	tags            string
	tagMode         string
	properties      string
	propertyMode    string
	flags       *flag.FlagSet
}

func NewCLI() *CLI {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	cli := &CLI{
		configPath: filepath.Join(homeDir, ".config", "godis"),
		flags:      flag.NewFlagSet("godis", flag.ExitOnError),
	}

	return cli
}

func (c *CLI) parseFlags(args []string) error {
	// Define flags with long and short versions
	c.flags.StringVar(&c.configFile, "config", "default.yaml", "Config file to use")
	c.flags.StringVar(&c.configFile, "c", "default.yaml", "Config file to use (shorthand)")
	
	c.flags.StringVar(&c.token, "token", "", "Discord bot token")
	c.flags.StringVar(&c.token, "t", "", "Discord bot token (shorthand)")
	
	c.flags.StringVar(&c.channelID, "channel", "", "Discord channel ID")
	c.flags.StringVar(&c.channelID, "ch", "", "Discord channel ID (shorthand)")
	
	c.flags.StringVar(&c.serverID, "server", "", "Discord server ID")
	c.flags.StringVar(&c.serverID, "s", "", "Discord server ID (shorthand)")
	
	c.flags.StringVar(&c.username, "username", "", "Bot username")
	c.flags.StringVar(&c.username, "u", "", "Bot username (shorthand)")

	c.flags.StringVar(&c.tags, "tags", "", "Comma-separated tags")
	c.flags.StringVar(&c.tagMode, "tag-mode", "merge", "Tag handling mode (merge|replace)")
	
	c.flags.StringVar(&c.properties, "properties", "", "Properties in key:value;key2:value2 format")
	c.flags.StringVar(&c.propertyMode, "property-mode", "merge", "Property handling mode (merge|replace)")


	return c.flags.Parse(args)
}

func (c *CLI) parseTags(tagStr string) []string {
	if tagStr == "" {
		return nil
	}
	tags := strings.Split(tagStr, ",")
	// Trim spaces from each tag
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}

func (c *CLI) parseProperties(propStr string) map[string]string {
	props := make(map[string]string)
	if propStr == "" {
		return props
	}
	
	pairs := strings.Split(propStr, ";")
	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), ":")
		if len(kv) == 2 {
			props[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return props
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

	if c.tags != "" {
		newTags := c.parseTags(c.tags)
		if c.tagMode == string(ModeReplace) {
			c.config.Tags = newTags
		} else { // merge mode
			// Create a map for deduplication
			tagMap := make(map[string]bool)
			for _, t := range c.config.Tags {
				tagMap[t] = true
			}
			for _, t := range newTags {
				tagMap[t] = true
			}
			// Convert back to slice
			c.config.Tags = make([]string, 0, len(tagMap))
			for t := range tagMap {
				c.config.Tags = append(c.config.Tags, t)
			}
		}
	}

	// Handle properties
	if c.properties != "" {
		newProps := c.parseProperties(c.properties)
		if c.propertyMode == string(ModeReplace) {
			c.config.Properties = newProps
		} else { // merge mode
			if c.config.Properties == nil {
				c.config.Properties = make(map[string]string)
			}
			for k, v := range newProps {
				c.config.Properties[k] = v
			}
		}
	}


}

func main() {
		cli := NewCLI()
		if err := cli.parseFlags(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}

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

