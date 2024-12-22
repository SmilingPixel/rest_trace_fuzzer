#!/bin/bash

# Set the output directory and binary name
OUTPUT_DIR="bin"
BINARY_NAME="api-fuzzer"

# Ensure the script is run from the root directory of the project
if [ ! -f "go.mod" ]; then
    echo "Error: Script must be run from the root directory of the project."
    exit 1
fi

# Create the output directory if it doesn't exist
mkdir -p $OUTPUT_DIR

# Get terminal dimensions
HEIGHT=$(tput lines)
WIDTH=$(tput cols)

# Build the project
{
    go build -v -x -o $OUTPUT_DIR/$BINARY_NAME ./cmd/api-fuzzer
} 2>&1 | pv -l -s $(go list ./... | wc -l) | dialog --progressbox "Building the project..." $HEIGHT $WIDTH

# Check if the build was successful
if [ $? -eq 0 ]; then
    echo "Build successful! Binary is located at $OUTPUT_DIR/$BINARY_NAME"
else
    echo "Build failed!"
    exit 1
fi