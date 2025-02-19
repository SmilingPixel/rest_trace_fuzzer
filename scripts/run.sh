#!/bin/bash

# Check if the correct number of arguments are provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <EXECUTABLE_PATH>"
    exit 1
fi

# Set the binary path from command line arguments
EXECUTABLE_PATH="$1"

# Ensure the binary exists
if [ ! -f "$EXECUTABLE_PATH" ]; then
    echo "Error: Binary not found. Please build the project first."
    exit 1
fi


# Arguments
# The configuration file
CONFIG_FILE="./config/my_config.json"

# Run the binary
echo "Running the binary..."
$EXECUTABLE_PATH \
    --config-file $CONFIG_FILE
