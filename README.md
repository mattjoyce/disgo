# Disgo

Disgo is a simple command-line tool that sends data from stdin to a Discord channel. It's designed to be easy to use in Unix pipelines and shell scripts.

## Features

- Send data from stdin to a Discord channel
- Create Discord threads for message groups
- Split long messages into smaller chunks
- Add tags and properties to organize messages
- Passthrough mode to echo stdin to stdout while also sending to Discord
- Supports YAML configuration files

## Installation

```bash
go install github.com/yourusername/disgo@latest
```

## Quick Start

```bash
# Send output from a command to Discord
echo "Hello from disgo!" | disgo --token "your-discord-bot-token" --channel "your-channel-id"

# Pipe log output to Discord
tail -f /var/log/application.log | disgo --thread "Application Logs"

# Send with tags for organization
cat report.txt | disgo --tags "report,daily" --thread "Daily Reports"
```

## Configuration

Disgo can be configured via command line flags or a YAML configuration file. Configuration files are stored in `~/.config/disgo/` with the extension `.yaml`.

### Example Configuration File

Create a file `~/.config/disgo/default.yaml`:

```yaml
token: "your-discord-bot-token"
channel_id: "your-channel-id"
server_id: "your-server-id"
username: "disgo-bot"
tags: ["log", "server"]
tag_mode: "merge"
properties:
  env: "production"
  region: "us-west"
property_mode: "merge"
debug: false
max_message_size: 2000
message_mode: "serialize"
thread_name: ""
passthrough: false
```

### Command Line Options

```
--config string        Config name to use (stored in ~/.config/disgo/NAME.yaml) (default "default")
-c string              Config name to use (shorthand) (default "default")
--token string         Discord bot token
-t string              Discord bot token (shorthand)
--channel string       Discord channel ID
-ch string             Discord channel ID (shorthand)
--server string        Discord server ID
-s string              Discord server ID (shorthand)
--username string      Bot username
-u string              Bot username (shorthand)
--tags string          Comma-separated tags
--tag-mode string      Tag handling mode (merge|replace) (default "merge")
--properties string    Properties in key:value;key2:value2 format
--property-mode string Property handling mode (merge|replace) (default "merge")
--debug                Enable debug logging
--passthrough          Echo stdin to stdout
--max-size int         Maximum message size (default 2000)
--message-mode string  Message handling mode (serialize|truncate) (default "serialize")
--thread string        Create thread with given name for messages
```

## Using with Logs

Disgo is great for sending log output to Discord. Here are a few common patterns:

```bash
# Stream a log file
tail -f /var/log/nginx/access.log | disgo --thread "Nginx Access Logs"

# Filter logs and send only errors
grep "ERROR" /var/log/application.log | disgo --tags "error,critical"

# Send periodic reports
crontab -e
# Add: 0 8 * * * cat /var/log/daily-report.txt | disgo -c reports --thread "$(date +\%Y-\%m-\%d) Report"
```

## License

MIT