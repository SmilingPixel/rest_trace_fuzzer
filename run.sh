#!/bin/bash

# Set the output directory and binary name
OUTPUT_DIR="bin"
BINARY_NAME="api-fuzzer"

# Ensure the binary exists
if [ ! -f "$OUTPUT_DIR/$BINARY_NAME" ]; then
    echo "Error: Binary not found. Please build the project first."
    exit 1
fi

# Run the binary
echo "Running the binary..."
$OUTPUT_DIR/$BINARY_NAME
