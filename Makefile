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
	bash scripts/run.sh "$(BUILD_DIR)/$(BINARY)" "$(CONFIG_FILE)"

# Example usage:
# make run CONFIG_FILE=./config/custom_config.json

# Clean the program output
.PHONY: clean-output
clean-output:
	bash scripts/clean_output.sh

# Clean the build
.PHONY: clean-build
clean-build:
	bash scripts/clean_build.sh "$(BUILD_DIR)/$(BINARY)"

# Clean both output and build
.PHONY: clean
clean: clean-output clean-build
