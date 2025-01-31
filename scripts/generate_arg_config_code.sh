#!/bin/bash

# This script generates the argument configuration code using a Python script
# and then formats the generated Go code.

# Path to the Python script that generates the argument configuration
PYTHON_SCRIPT_PATH="./internal/config/arg_config_generate.py"

# Path to the module where the generated Go code will be located
MODULE_PATH="./internal/config"

# Call the Python script to generate the argument configuration
python3 $PYTHON_SCRIPT_PATH $MODULE_PATH

# Format the Go code in the specified directory
go fmt $MODULE_PATH