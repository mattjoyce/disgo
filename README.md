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
go install github.com/mattjoyce/disgo@latest
```

Or clone and build:

```bash
git clone https://github.com/mattjoyce/disgo.git
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

## Integration Examples

### Python Logging Handler

You can use Disgo as a logging handler for Python applications to send logs directly to Discord. Here's a working example:

```python
import logging
import subprocess
import sys

class DisgoHandler(logging.Handler):
    def __init__(self, config='default', thread=None, tags=None, level=logging.WARNING):
        super().__init__(level=level)
        self.config = config
        self.thread = thread
        self.tags = tags
        self.formatter = logging.Formatter('%(levelname)s - %(name)s - %(message)s')
        
    def emit(self, record):
        try:
            msg = self.format(record)
            cmd = ['disgo', '--config', self.config]
            
            if self.thread:
                cmd.extend(['--thread', self.thread])
                
            if self.tags:
                cmd.extend(['--tags', self.tags])
                
            # Add level as a tag for easier filtering
            if record.levelno >= logging.ERROR:
                cmd.extend(['--tags', 'error'])
                
            process = subprocess.Popen(
                cmd,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            
            process.communicate(input=msg)
            
        except Exception:
            self.handleError(record)

# Example usage
if __name__ == "__main__":
    # Configure the root logger
    logger = logging.getLogger()
    logger.setLevel(logging.INFO)
    
    # Console handler for all logs
    console = logging.StreamHandler(sys.stdout)
    console.setLevel(logging.INFO)
    logger.addHandler(console)
    
    # Disgo handler for warning and above
    disgo_handler = DisgoHandler(
        thread="Application Logs",
        tags="python,app",
        level=logging.WARNING
    )
    logger.addHandler(disgo_handler)
    
    # Test log messages
    logger.info("This is an info message - not sent to Discord")
    logger.warning("This is a warning - sent to Discord")
    logger.error("This is an error - sent to Discord with error tag")
    
    try:
        x = 1 / 0
    except Exception as e:
        logger.exception("Caught an exception")
```

This implementation:
- Creates a custom logging handler that pipes log messages to Disgo
- Only sends WARNING level and above to Discord by default
- Automatically adds the "error" tag for ERROR and above
- Supports custom thread names and tags
- Formats messages with level, logger name, and message

### Go Logging Integration

You can also use Disgo with Go's logging capabilities by creating a custom writer:

```go
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

// DisgoWriter implements io.Writer to pipe logs to disgo
type DisgoWriter struct {
	Config     string
	ThreadName string
	Tags       string
	Level      string // Log level for conditional tagging
}

// Write satisfies io.Writer interface
func (w *DisgoWriter) Write(p []byte) (n int, err error) {
	cmd := exec.Command("disgo", "--config", w.Config)
	
	if w.ThreadName != "" {
		cmd.Args = append(cmd.Args, "--thread", w.ThreadName)
	}
	
	// Base tags
	tags := w.Tags
	
	// Add level-based tags if applicable
	if w.Level == "ERROR" {
		if tags != "" {
			tags += ","
		}
		tags += "error,critical"
	}
	
	if tags != "" {
		cmd.Args = append(cmd.Args, "--tags", tags)
	}
	
	// Create stdin pipe to send log content
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}
	
	// Start the command before writing to stdin
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	
	// Write to stdin and close
	n, err = stdin.Write(p)
	stdin.Close()
	
	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return n, err
	}
	
	return n, nil
}

func main() {
	// Create multi-writers for each level
	errorWriter := io.MultiWriter(
		os.Stdout,
		&DisgoWriter{
			Config:     "default",
			ThreadName: "Go Application Logs",
			Tags:       "golang,app",
			Level:      "ERROR",
		},
	)
	
	// Create logger with level prefix
	errorLogger := log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime)
	
	// Log error messages
	errorLogger.Println("This error will be sent to Discord with the error tag")
}
```

This Go implementation:
- Creates a custom io.Writer that pipes log messages to Disgo
- Can be configured with thread names and tags
- Can be used with Go's standard logging package or other loggers
- Supports tagging based on log level
- Can be combined with existing writers using io.MultiWriter

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

GPL-3.0