#!/bin/bash

# Set the output directory and binary name
OUTPUT_DIR="bin"
BINARY_NAME="api-fuzzer"

# Ensure the binary exists
if [ ! -f "$OUTPUT_DIR/$BINARY_NAME" ]; then
    echo "Error: Binary not found. Please build the project first."
    exit 1
fi


# Arguments
# 1. The input OpenAPI file
OAS_FILE="../openapi/otel_demo_swagger.json"
# 2. The RESTler dependency file
RESTLER_DEPENDENCY_FILE="../restler_bin/restler/Compile/dependencies.json"
# 3. Dependency file type
DEPENDENCY_FILE_TYPE="Restler"

# Run the binary
echo "Running the binary..."
$OUTPUT_DIR/$BINARY_NAME \
    --openapi-spec $OAS_FILE \
    --dependency-file $RESTLER_DEPENDENCY_FILE \
    --dependency-file-type $DEPENDENCY_FILE_TYPE
