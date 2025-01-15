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
OAS_FILE="../openapi/otel_demo/system_swagger.json"
# 2. The RESTler dependency file
RESTLER_DEPENDENCY_FILE="../restler_bin/restler/Compile/dependencies.json"
# 3. Dependency file type
DEPENDENCY_FILE_TYPE="Restler"
# 4. Fuzzer type
FUZZER_TYPE="Basic"
# 5. Fuzzing budget
FUZZING_BUDGET="30s"
# 6. Server base URL
SERVER_BASE_URL="http://www.example.com"
# 7. Internal service OpenAPI map file
INTERNAL_SERVICE_OAS_MAP_FILE="../openapi/otel_demo/service2oas.json"
# 8. Output directory
OUTPUT_DIR="./output"

echo "INTERNAL_SERVICE_OAS_FILES: $INTERNAL_SERVICE_OAS_FILES"

# Run the binary
echo "Running the binary..."
$OUTPUT_DIR/$BINARY_NAME \
    --openapi-spec $OAS_FILE \
    --fuzzer-type $FUZZER_TYPE \
    --fuzzer-budget $FUZZING_BUDGET \
    --server-base-url $SERVER_BASE_URL \
    --internal-service-openapi-spec-map $INTERNAL_SERVICE_OAS_MAP_FILE \
    --output-dir $OUTPUT_DIR
