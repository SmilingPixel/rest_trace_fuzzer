#!/bin/bash

# Ensure the script is run from the root directory of the project
if [ ! -f "go.mod" ]; then
    echo "Error: Script must be run from the root directory of the project."
    exit 1
fi

# Check if the correct number of arguments are provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <output_directory> <binary_name>"
    exit 1
fi

# Set the output directory and binary name from command line arguments
OUTPUT_DIR="$1"
BINARY_NAME="$2"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Build the project
echo "Building the project..."
go build -v -x -o "$OUTPUT_DIR/$BINARY_NAME" ./cmd/api-fuzzer

# Check if the build was successful
BUILD_STATUS=$?
if [ $BUILD_STATUS -eq 0 ]; then
    echo "Build successful! Binary is located at $OUTPUT_DIR/$BINARY_NAME"
else
    echo "Build failed with status code $BUILD_STATUS!"
    exit 1
fi