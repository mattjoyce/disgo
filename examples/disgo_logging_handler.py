#!/usr/bin/env python3
"""
Disgo Python Logging Handler Example

This example demonstrates how to create a custom logging handler
that sends log messages to Discord using Disgo.
"""

import logging
import subprocess
import sys
import time


class DisgoHandler(logging.Handler):
    """
    Custom logging handler that pipes log messages to Disgo.
    Only sends messages at or above the configured level.
    """

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
                cmd.extend(['--tags', 'error,critical'])
            elif record.levelno >= logging.WARNING:
                cmd.extend(['--tags', 'warning'])
                
            process = subprocess.Popen(
                cmd,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            
            process.communicate(input=msg)
            
        except Exception as e:
            self.handleError(record)


def configure_logging():
    """Configure logging with both console and Discord handlers."""
    # Get the root logger
    logger = logging.getLogger()
    logger.setLevel(logging.INFO)
    
    # Console handler for all logs
    console = logging.StreamHandler(sys.stdout)
    console.setLevel(logging.INFO)
    console.setFormatter(logging.Formatter('%(asctime)s - %(levelname)s - %(name)s - %(message)s'))
    logger.addHandler(console)
    
    # Disgo handler for warning and above
    disgo_handler = DisgoHandler(
        thread="Application Logs",
        tags="python,example",
        level=logging.WARNING
    )
    disgo_handler.setFormatter(logging.Formatter('%(asctime)s - %(levelname)s - %(name)s - %(message)s'))
    logger.addHandler(disgo_handler)
    
    return logger


def simulate_activity(logger):
    """Simulate application activity with various log levels."""
    logger.info("Starting simulation")
    
    # Generate some info messages
    for i in range(3):
        logger.info(f"Processing item {i}")
        time.sleep(1)
    
    # Generate a warning
    logger.warning("Resource usage is high (80%)")
    time.sleep(1)
    
    # Generate an error
    logger.error("Failed to connect to external API")
    time.sleep(1)
    
    # Simulate an exception
    try:
        result = 100 / 0
    except Exception as e:
        logger.exception("An unexpected error occurred")
    
    logger.info("Simulation complete")


if __name__ == "__main__":
    logger = configure_logging()
    
    # Add app logger for more specific context
    app_logger = logging.getLogger("ExampleApp")
    
    # Run the simulation
    simulate_activity(app_logger)