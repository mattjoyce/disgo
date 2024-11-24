package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v3"
)

type MergeMode string
const (
	ModeReplace MergeMode = "replace"
	ModeMerge   MergeMode = "merge"
)

const (
	DefaultMaxMessageSize = 2000
	ModeSerialize = "serialize"
	ModeTruncate = "truncate"
)


// Config holds all configuration options
type Config struct {
	Token         string            `yaml:"token"`
	ChannelID     string            `yaml:"channel_id"`
	ServerID      string            `yaml:"server_id"`
	Username      string            `yaml:"username"`
	Tags          []string          `yaml:"tags"`
	TagMode       string            `yaml:"tag_mode"`
	Properties    map[string]string `yaml:"properties"`
	PropertyMode  string            `yaml:"property_mode"`
	Debug         bool              `yaml:"debug"`
	MaxMessageSize int    `yaml:"max_message_size"`
  MessageMode    string `yaml:"message_mode"`
	Passthrough bool `yaml:"passthrough"`
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
	debug    bool
	passthrough bool
	stdinData   []byte
  maxMessageSize int
  messageMode    string
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

	c.flags.BoolVar(&c.debug, "debug", false, "Enable debug logging")

	c.flags.BoolVar(&c.passthrough, "passthrough", false, "Echo stdin to stdout")

	c.flags.IntVar(&c.maxMessageSize, "max-size", DefaultMaxMessageSize, "Maximum message size")
	c.flags.StringVar(&c.messageMode, "message-mode", ModeSerialize, "Message handling mode (serialize|truncate)")

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
	data, err := os.ReadFile(configFile)
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
			Token:        "",
			ChannelID:    "",
			ServerID:     "",
			Username:     "godis-bot",
			Tags:         []string{},
			TagMode:      "merge",
			Properties:   map[string]string{},
			PropertyMode: "merge",
			Debug:        false,
			MaxMessageSize: DefaultMaxMessageSize,
			MessageMode:    ModeSerialize,
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
			return fmt.Errorf("failed to write default config: %w", err)
	}

	c.config = defaultConfig
	return nil
}

func (c *CLI) mergeFlags() {
	// Existing merges
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
	if c.debug {
			c.config.Debug = true
	}

	if c.passthrough {
		c.config.Passthrough = true
	}
	// Mode settings
	if c.tagMode != "" {
			c.config.TagMode = c.tagMode
	}
	if c.propertyMode != "" {
			c.config.PropertyMode = c.propertyMode
	}

	if c.maxMessageSize != DefaultMaxMessageSize {
		c.config.MaxMessageSize = c.maxMessageSize
	}
	if c.messageMode != "" {
			c.config.MessageMode = c.messageMode
	}

	// Handle tags with configured mode
	if c.tags != "" {
			newTags := c.parseTags(c.tags)
			if c.config.TagMode == string(ModeReplace) {
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

	// Handle properties with configured mode
	if c.properties != "" {
			newProps := c.parseProperties(c.properties)
			if c.config.PropertyMode == string(ModeReplace) {
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

func (c *CLI) readStdin() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
			return fmt.Errorf("error checking stdin: %w", err)
	}

	// Check if we have data on stdin
	if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
					return fmt.Errorf("error reading stdin: %w", err)
			}
			c.stdinData = data
	}
	return nil
}

func (c *CLI) sendToDiscord() error {
	if c.config.Token == "" {
			return fmt.Errorf("discord token not configured")
	}
	if c.config.ChannelID == "" {
			return fmt.Errorf("discord channel ID not configured")
	}

	// Ensure token has "Bot " prefix if not already present
	token := c.config.Token
	if !strings.HasPrefix(token, "Bot ") {
			token = "Bot " + token
	}

	if c.config.Debug {
			log.Printf("Creating Discord session with token length: %d", len(c.config.Token))
			log.Printf("Using channel ID: %s", c.config.ChannelID)
	}

	// Create Discord session
	discord, err := discordgo.New(token)
	if err != nil {
			return fmt.Errorf("error creating Discord session: %w", err)
	}
	defer discord.Close()

	// Verify the token by making a test API call
	_, err = discord.User("@me")
	if err != nil {
			if strings.Contains(err.Error(), "401") {
					return fmt.Errorf("invalid Discord token. Please check your configuration")
			}
			return fmt.Errorf("error verifying Discord token: %w", err)
	}

	if c.config.Debug {
			log.Printf("Message length: %d", len(c.stdinData))
	}

	if len(c.stdinData) > 0 {
		content := string(c.stdinData)
		messages := c.splitMessage(content)

		if c.config.Debug {
				log.Printf("Splitting content of length %d into %d messages", len(content), len(messages))
		}

		for i, msg := range messages {
				if c.config.Debug {
						log.Printf("Sending message part %d/%d (length: %d)", i+1, len(messages), len(msg))
				}

				_, err = discord.ChannelMessageSend(c.config.ChannelID, msg)
				if err != nil {
						return fmt.Errorf("error sending message part %d/%d: %w", i+1, len(messages), err)
				}
		}
	}

	return nil
	}

func (c *CLI) splitMessage(content string) []string {
	if c.config.Debug {
			log.Printf("Splitting message of length %d", len(content))
	}

	if len(content) <= c.config.MaxMessageSize {
			if c.config.Debug {
					log.Printf("Message is less than max size, not splitting")
			}
			return []string{content}
	}

	switch c.config.MessageMode {
	case ModeTruncate:
			return []string{content[:c.config.MaxMessageSize]}
	case ModeSerialize:
			var messages []string
			remaining := content
			for len(remaining) > 0 {
					splitAt := c.config.MaxMessageSize
					if len(remaining) < splitAt {
							splitAt = len(remaining)
					}

					// Try to split at newline if possible
					if splitAt < len(remaining) {
							lastNewline := strings.LastIndex(remaining[:splitAt], "\n")
							if lastNewline > 0 {
									splitAt = lastNewline + 1
							}
					}

					messages = append(messages, remaining[:splitAt])
					remaining = remaining[splitAt:]
			}
			return messages
	default:
			// Default to truncate if invalid mode
			return []string{content[:c.config.MaxMessageSize]}
	}
}

func main() {
	cli := NewCLI()
	if err := cli.parseFlags(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
	}

	if err := cli.readStdin(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
	}

	err := cli.loadConfig()
	if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
	}

	cli.mergeFlags()

	
	if cli.config.Debug {
		  log.Printf("Starting godis...")
			log.Printf("Debug logging enabled")
			log.Printf("Using configuration:")
			log.Printf("Token: %s", cli.config.Token)
			log.Printf("Channel ID: %s", cli.config.ChannelID)
			log.Printf("Server ID: %s", cli.config.ServerID)
			log.Printf("Username: %s", cli.config.Username)
			log.Printf("Max message size: %d", cli.config.MaxMessageSize)
			log.Printf("Message mode: %s", cli.config.MessageMode)
			log.Printf("Tags: %v", cli.config.Tags)
			log.Printf("Properties: %v", cli.config.Properties)
			log.Printf("Passthrough: %v", cli.config.Passthrough)
	}

	// Handle passthrough if enabled
	if cli.config.Passthrough && len(cli.stdinData) > 0 {
			os.Stdout.Write(cli.stdinData)
	}

	if err := cli.sendToDiscord(); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending to Discord: %v\n", err)
		os.Exit(1)
	}

}

