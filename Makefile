# Makefile for microservice_test/rest_trace_fuzzer

# Generate configuration code
.PHONY: generate
generate:
	bash scripts/generate_arg_config_code.sh

# Build the project
.PHONY: build
build:
	bash scripts/build.sh

# Run the project
.PHONY: run
run:
	bash scripts/run.sh

# Clean the program output
.PHONY: clean
clean:
	bash scripts/clean_output.sh
