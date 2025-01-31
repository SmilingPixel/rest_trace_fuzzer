#!/bin/bash

# Set the output directory and binary name
BIN_OUTPUT_DIR="bin"
BINARY_NAME="api-fuzzer"

# Ensure the binary exists
if [ ! -f "$BIN_OUTPUT_DIR/$BINARY_NAME" ]; then
    echo "Error: Binary not found. Please build the project first."
    exit 1
fi


# Arguments
# The configuration file
CONFIG_FILE="./config/config.json"

# Run the binary
echo "Running the binary..."
$BIN_OUTPUT_DIR/$BINARY_NAME \
    --config-file $CONFIG_FILE
