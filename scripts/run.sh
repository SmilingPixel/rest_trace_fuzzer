#!/bin/bash

# Check if the correct number of arguments are provided
if [ "$#" -lt 1 ]; then
    echo "Usage: $0 <EXECUTABLE_PATH> [CONFIG_FILE]"
    exit 1
fi

# Set the binary path from command line arguments
EXECUTABLE_PATH="$1"

# Set the configuration file, default to "./config/config.json" if not provided
CONFIG_FILE="${2:-./config/config.json}"

# Ensure the binary exists
if [ ! -f "$EXECUTABLE_PATH" ]; then
    echo "Error: Binary not found. Please build the project first."
    exit 1
fi

# Run the binary
echo "Running the binary with config file: $CONFIG_FILE..."
$EXECUTABLE_PATH \
    --config-file "$CONFIG_FILE"
