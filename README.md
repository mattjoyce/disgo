# Disgo

A simple CLI tool for sending messages to Discord channels from the command line. Perfect for integrating Discord notifications into scripts and pipelines.

## Features

- Pipe text content directly to Discord channels
- Support for configuration files
- Thread creation for long messages
- Message handling modes for large content
- Debug logging
- Passthrough mode for testing

## Installation

```bash
go install github.com/yourusername/disgo@latest
```

Or clone and build:

```bash
git clone https://github.com/yourusername/disgo.git
cd disgo
go build -o disgo
```

## Setup

1. Create a Discord bot and get its token ([Discord Developer Portal](https://discord.com/developers/applications))
2. Get your channel ID (Enable Developer Mode in Discord, right-click channel, Copy ID)
3. Create or edit the config file at `~/.config/disgo/default.yaml`:

```yaml
token: "your-bot-token"
channel_id: "your-channel-id"
username: "disgo-bot"
```

## Usage

Basic usage:
```bash
# Send a message
echo "Hello Discord!" | disgo

# Send with debug information
echo "Hello Discord!" | disgo --debug

# Use passthrough mode for testing
echo "Test message" | disgo --passthrough

# Create a thread for long content
cat longfile.txt | disgo --thread "Long Content Thread"

# Use a different config file
echo "Using alt config" | disgo --config test1
```

## Configuration

Default configuration location: `~/.config/disgo/default.yaml`

```yaml
token: ""                  # Discord bot token
channel_id: ""            # Target channel ID
server_id: ""             # Server ID (optional)
username: "disgo-bot"     # Bot username
debug: false              # Enable debug logging
max_message_size: 2000    # Maximum message size
message_mode: "serialize" # Message handling mode (serialize|truncate)
thread_name: ""          # Default thread name (optional)
```

Multiple configuration files can be used by placing them in the `~/.config/disgo/` directory with a `.yaml` extension.

## Message Handling

Long messages (>2000 characters) are handled in two ways:

- `serialize`: Splits the message into multiple parts (default)
- `truncate`: Cuts off at the maximum length

When using threads, the first message will be a thread notification, and the content will be posted within the thread.

## Command Line Options

```
  -c, --config string       Config name to use (without .yaml extension) (default "default")
      --debug              Enable debug logging
      --max-size int       Maximum message size (default 2000)
      --message-mode string Message handling mode (serialize|truncate) (default "serialize")
      --passthrough        Echo stdin to stdout
      --thread string      Create thread with given name for messages
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT License](LICENSE)
