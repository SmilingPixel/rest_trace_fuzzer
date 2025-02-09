# Makefile for microservice_test/rest_trace_fuzzer

# Define build target path
BUILD_DIR := bin
BINARY := api-fuzzer

# Generate configuration code
.PHONY: generate
generate:
	bash scripts/generate_arg_config_code.sh

# Build the project
.PHONY: build
build:
	bash scripts/build.sh "$(BUILD_DIR)" "$(BINARY)"

# Run the project
.PHONY: run
run:
	bash scripts/run.sh "$(BUILD_DIR)/$(BINARY)"

# Clean the program output
.PHONY: clean
clean:
	bash scripts/clean_output.sh
	bash scripts/clean_build.sh "$(BUILD_DIR)/$(BINARY)"
